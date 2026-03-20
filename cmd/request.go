package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"github.com/smilemilks2021/easy-web/internal/auth"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
)

func init() {
	cmd := &cobra.Command{
		Use:   "request",
		Short: "Make an authenticated HTTP request",
		RunE:  runRequest,
	}

	// existing flags
	cmd.Flags().StringP("method", "X", "GET", "HTTP method")
	cmd.Flags().StringP("data", "d", "", "Request body")
	cmd.Flags().StringArrayP("header", "H", nil, "Extra header (Key: Value)")

	// output format
	cmd.Flags().StringP("output", "o", "json", "Output format: json | raw | headers")

	// verbose
	cmd.Flags().BoolP("verbose", "v", false, "Show request/response headers")

	// retry
	cmd.Flags().Int("retry", 0, "Retry count on 5xx or network error")
	cmd.Flags().Duration("retry-delay", time.Second, "Delay between retries")

	// proxy
	cmd.Flags().String("proxy", "", "Proxy URL (e.g. http://127.0.0.1:8080)")

	// form / file body
	cmd.Flags().StringArray("form", nil, "Form field key=value (sets application/x-www-form-urlencoded)")
	cmd.Flags().StringArray("file", nil, "Multipart file field@/path (sets multipart/form-data)")

	// jq filter
	cmd.Flags().String("jq", "", "jq expression to filter JSON response")

	// auth mode (forwarded from existing flag in auth.go; duplicated here for --mode support)
	cmd.Flags().String("mode", "", "Auth mode override")
	cmd.Flags().StringP("url", "u", "", "Target URL")

	rootCmd.AddCommand(cmd)
}

func runRequest(cmd *cobra.Command, _ []string) error {
	targetURL, _ := cmd.Flags().GetString("url")
	if targetURL == "" {
		return fmt.Errorf("--url is required")
	}

	method, _ := cmd.Flags().GetString("method")
	body, _ := cmd.Flags().GetString("data")
	hdrs, _ := cmd.Flags().GetStringArray("header")
	outputFmt, _ := cmd.Flags().GetString("output")
	verbose, _ := cmd.Flags().GetBool("verbose")
	retryCount, _ := cmd.Flags().GetInt("retry")
	retryDelay, _ := cmd.Flags().GetDuration("retry-delay")
	proxyURL, _ := cmd.Flags().GetString("proxy")
	formFields, _ := cmd.Flags().GetStringArray("form")
	fileFields, _ := cmd.Flags().GetStringArray("file")
	jqExpr, _ := cmd.Flags().GetString("jq")
	mode, _ := cmd.Flags().GetString("mode")

	// --- auth / cookies ---
	store := cookie.NewCache(config.CacheDir())
	entries, err := store.Load(parseHost(targetURL))
	if err != nil {
		return fmt.Errorf("reading cookie cache: %w", err)
	}
	if len(entries) == 0 {
		result, err := auth.Resolve(targetURL, auth.Options{Mode: mode})
		if err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		entries = result.Cookies
	}

	// --- extra headers from -H ---
	extra := map[string]string{}
	for _, h := range hdrs {
		k, v, ok := strings.Cut(h, ":")
		if ok {
			extra[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}

	// --- build body + override content-type for --form / --file ---
	callHeaders := map[string]string{}
	var finalBody string

	switch {
	case len(fileFields) > 0:
		// multipart/form-data
		buf := &bytes.Buffer{}
		w := multipart.NewWriter(buf)
		// add plain form fields first
		for _, f := range formFields {
			k, v, _ := strings.Cut(f, "=")
			_ = w.WriteField(k, v)
		}
		// add file fields
		for _, f := range fileFields {
			fieldName, filePath, ok := strings.Cut(f, "@")
			if !ok {
				return fmt.Errorf("--file must be in field@/path format, got: %s", f)
			}
			part, err := w.CreateFormFile(fieldName, filePath)
			if err != nil {
				return fmt.Errorf("create form file: %w", err)
			}
			fh, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("open file %s: %w", filePath, err)
			}
			_, copyErr := io.Copy(part, fh)
			fh.Close()
			if copyErr != nil {
				return fmt.Errorf("read file %s: %w", filePath, copyErr)
			}
		}
		w.Close()
		finalBody = buf.String()
		callHeaders["Content-Type"] = w.FormDataContentType()

	case len(formFields) > 0:
		// application/x-www-form-urlencoded
		vals := url.Values{}
		for _, f := range formFields {
			k, v, _ := strings.Cut(f, "=")
			vals.Add(k, v)
		}
		finalBody = vals.Encode()
		callHeaders["Content-Type"] = "application/x-www-form-urlencoded"

	default:
		finalBody = body
	}

	// merge callHeaders into extra so Do() receives them via callHeaders param
	opts := request.Options{
		Retry:      retryCount,
		RetryDelay: retryDelay,
		ProxyURL:   proxyURL,
	}
	c := request.NewClient(entries, extra, opts)

	// --- verbose: print request info ---
	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "> %s %s\n", method, targetURL)
		for k, v := range extra {
			fmt.Fprintf(cmd.OutOrStdout(), "> %s: %s\n", k, v)
		}
		if len(entries) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "> Cookie: %s\n", cookie.CookieHeader(entries))
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	resp, err := c.Do(method, targetURL, finalBody, callHeaders)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// --- verbose: print response headers ---
	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "< HTTP %d %s\n", resp.StatusCode, resp.Status[strings.Index(resp.Status, " ")+1:])
		for k, vs := range resp.Header {
			for _, v := range vs {
				fmt.Fprintf(cmd.OutOrStdout(), "< %s: %s\n", k, v)
			}
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	out := cmd.OutOrStdout()

	switch outputFmt {
	case "headers":
		for k, vs := range resp.Header {
			for _, v := range vs {
				fmt.Fprintf(out, "%s: %s\n", k, v)
			}
		}
		return nil

	case "raw":
		fmt.Fprintf(out, "HTTP %d\n", resp.StatusCode)
		io.Copy(out, resp.Body)
		fmt.Fprintln(out)
		return nil

	default: // "json" or anything else → pretty JSON with highlight
		fmt.Fprintf(out, "HTTP %d\n", resp.StatusCode)
		raw, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return readErr
		}

		// apply jq filter if requested
		if jqExpr != "" {
			if filtered, ok := applyJQ(raw, jqExpr); ok {
				raw = filtered
			}
		}

		// try to pretty-print as JSON
		var prettyBuf bytes.Buffer
		if json.Valid(raw) {
			if err := json.Indent(&prettyBuf, raw, "", "  "); err == nil {
				fmt.Fprintln(out, colorizeJSON(prettyBuf.Bytes()))
				return nil
			}
		}
		// fallback: raw output
		out.Write(raw)
		fmt.Fprintln(out)
		return nil
	}
}

// colorizeJSON applies ANSI colours to an already-indented JSON byte slice.
// It processes the text character-by-character using a tiny state machine so
// it never needs an external dependency.
func colorizeJSON(data []byte) string {
	const (
		colorReset  = "\033[0m"
		colorBlue   = "\033[34m" // JSON string key
		colorGreen  = "\033[32m" // JSON string value
		colorYellow = "\033[33m" // number / bool / null
	)

	var sb strings.Builder
	sb.Grow(len(data) * 2)

	// Simple state machine
	const (
		stateNormal = iota
		stateInString
		stateInEscape
	)
	state := stateNormal

	// We need to know whether the current string is a key or a value.
	// After a '{' or ',', the next string is a key.
	// After a ':', the next string is a value.
	nextStringIsKey := false

	// Track whether we just finished a colon (so the next token is a value).
	justColon := false

	i := 0
	for i < len(data) {
		ch := data[i]

		switch state {
		case stateNormal:
			switch ch {
			case '"':
				state = stateInString
				// Determine key vs value by scanning back over whitespace.
				// We track this via nextStringIsKey set when we see '{' or ',',
				// and clear it when we see ':'.
				if nextStringIsKey {
					sb.WriteString(colorBlue)
				} else {
					sb.WriteString(colorGreen)
				}
				sb.WriteByte(ch)
				justColon = false
			case '{', '[':
				sb.WriteByte(ch)
				nextStringIsKey = (ch == '{')
				justColon = false
			case '}', ']':
				sb.WriteByte(ch)
				nextStringIsKey = false
				justColon = false
			case ':':
				sb.WriteByte(ch)
				nextStringIsKey = false
				justColon = true
			case ',':
				sb.WriteByte(ch)
				// after comma: if we're inside an object the next token will be a key,
				// but we don't know here. We rely on peeking at the next non-space char.
				// Simpler approach: set a flag and decide when we actually write the token.
				justColon = false
				nextStringIsKey = false // will be set below after peeking
				// peek forward to decide if we're in an object or array
				j := i + 1
				for j < len(data) && (data[j] == ' ' || data[j] == '\n' || data[j] == '\r' || data[j] == '\t') {
					j++
				}
				if j < len(data) && data[j] == '"' {
					// could be key or string value; we need context. We set
					// nextStringIsKey=true only if the last open bracket was '{'.
					// Since we don't track a full stack, we use a heuristic: scan
					// backwards for the nearest unmatched '{' or '['.
					nextStringIsKey = lastOpenBracket(data[:i]) == '{'
				}
			default:
				// numbers, booleans, null, whitespace
				if ch == 't' || ch == 'f' || ch == 'n' || ch == '-' || (ch >= '0' && ch <= '9') {
					// colour the full token
					sb.WriteString(colorYellow)
					for i < len(data) && !isTokenEnd(data[i]) {
						sb.WriteByte(data[i])
						i++
					}
					sb.WriteString(colorReset)
					justColon = false
					continue
				}
				sb.WriteByte(ch)
			}
		case stateInString:
			sb.WriteByte(ch)
			switch ch {
			case '\\':
				state = stateInEscape
			case '"':
				sb.WriteString(colorReset)
				state = stateNormal
				// After a string that was a key, the colon comes next → nextStringIsKey=false
				// already handled by ':' case.
				// After a closing quote that was a value, next string (after comma) needs
				// to be a key if in an object → handled by ',' case above.
			}
		case stateInEscape:
			sb.WriteByte(ch)
			state = stateInString
		}
		i++
		_ = justColon // suppress unused warning; variable used indirectly
	}
	return sb.String()
}

// lastOpenBracket returns the last unmatched '{' or '[' scanning backwards in data.
func lastOpenBracket(data []byte) byte {
	depth := 0
	for i := len(data) - 1; i >= 0; i-- {
		switch data[i] {
		case '}', ']':
			depth++
		case '{', '[':
			if depth == 0 {
				return data[i]
			}
			depth--
		}
	}
	return 0
}

// isTokenEnd returns true for characters that end a JSON primitive token.
func isTokenEnd(ch byte) bool {
	return ch == ',' || ch == '}' || ch == ']' || ch == ':' || unicode.IsSpace(rune(ch))
}

// applyJQ runs a jq expression against raw JSON bytes and returns the
// marshalled result. Returns (nil, false) on any error.
func applyJQ(raw []byte, expr string) ([]byte, bool) {
	q, err := gojq.Parse(expr)
	if err != nil {
		return nil, false
	}
	var input interface{}
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, false
	}
	iter := q.Run(input)
	var results []interface{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, isErr := v.(error); isErr {
			_ = err
			break
		}
		results = append(results, v)
	}
	if len(results) == 0 {
		return nil, false
	}
	var out interface{}
	if len(results) == 1 {
		out = results[0]
	} else {
		out = results
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil, false
	}
	return b, true
}
