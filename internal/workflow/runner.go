package workflow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
	"gopkg.in/yaml.v3"
)

// Workflow is the top-level structure parsed from a workflow YAML file.
type Workflow struct {
	Name    string `yaml:"name"`
	BaseURL string `yaml:"base_url"`
	Steps   []Step `yaml:"steps"`
}

// Step describes a single HTTP request step.
type Step struct {
	Name    string            `yaml:"name"`
	Method  string            `yaml:"method"`
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Body    string            `yaml:"body"`
	Extract []Extract         `yaml:"extract"`
	Assert  []Assert          `yaml:"assert"`
}

// Extract defines a variable to capture from the response body using a jq expression.
type Extract struct {
	Var string `yaml:"var"`
	JQ  string `yaml:"jq"`
}

// Assert defines a single assertion on the response.
type Assert struct {
	Status int         `yaml:"status,omitempty"`
	JQ     string      `yaml:"jq,omitempty"`
	Equals interface{} `yaml:"equals,omitempty"`
}

// Runner executes a Workflow, threading variables across steps.
type Runner struct {
	cookies []*cookie.Entry
	vars    map[string]string
}

// NewRunner creates a Runner pre-loaded with cookies for the workflow's base domain.
func NewRunner(cookies []*cookie.Entry) *Runner {
	return &Runner{
		cookies: cookies,
		vars:    map[string]string{},
	}
}

// Load parses a workflow YAML file from disk.
func Load(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read workflow file: %w", err)
	}
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("parse workflow YAML: %w", err)
	}
	return &wf, nil
}

// Run executes all steps in the workflow, printing progress to stdout.
// It returns the number of assertion failures encountered.
func (r *Runner) Run(wf *Workflow) (int, error) {
	if wf.BaseURL != "" {
		r.vars["base_url"] = wf.BaseURL
	}

	total := len(wf.Steps)
	failures := 0

	for i, step := range wf.Steps {
		stepNum := i + 1

		// Resolve final URL (substitute variables, then handle relative vs absolute).
		resolvedURL := r.substitute(step.URL)
		if wf.BaseURL != "" && !isAbsoluteURL(resolvedURL) {
			resolvedURL = strings.TrimRight(wf.BaseURL, "/") + "/" + strings.TrimLeft(resolvedURL, "/")
		}

		method := step.Method
		if method == "" {
			method = "GET"
		}

		// Print step header.
		displayURL := resolvedURL
		if wf.BaseURL != "" && strings.HasPrefix(displayURL, wf.BaseURL) {
			displayURL = step.URL // show the original relative path for readability
		}
		fmt.Printf("[%d/%d] %s %s\n", stepNum, total, method, displayURL)

		// Substitute variables in body and headers.
		body := r.substitute(step.Body)
		headers := make(map[string]string, len(step.Headers))
		for k, v := range step.Headers {
			headers[k] = r.substitute(v)
		}

		// Execute the request.
		client := request.NewClient(r.cookies, headers)
		resp, err := client.Do(method, resolvedURL, body, nil)
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
			failures++
			continue
		}

		// Read response body.
		rawBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("  ERROR reading response body: %v\n", err)
			failures++
			continue
		}

		fmt.Printf("  \u2713 HTTP %d\n", resp.StatusCode)

		// Parse JSON response for assertions and extractions.
		var jsonData interface{}
		_ = json.Unmarshal(rawBody, &jsonData)

		// Run assertions.
		stepFailed := false
		for _, a := range step.Assert {
			if a.Status != 0 {
				if resp.StatusCode == a.Status {
					fmt.Printf("  \u2713 assert: status == %d\n", a.Status)
				} else {
					fmt.Printf("  \u2717 assert: status == %d (got %d)\n", a.Status, resp.StatusCode)
					stepFailed = true
				}
			}
			if a.JQ != "" {
				val, err := evalJQ(a.JQ, jsonData)
				if err != nil {
					fmt.Printf("  \u2717 assert: jq %q error: %v\n", a.JQ, err)
					stepFailed = true
					continue
				}
				if assertEquality(val, a.Equals) {
					fmt.Printf("  \u2713 assert: %s == %v\n", a.JQ, a.Equals)
				} else {
					fmt.Printf("  \u2717 assert: %s == %v (got %v)\n", a.JQ, a.Equals, val)
					stepFailed = true
				}
			}
		}
		if stepFailed {
			failures++
		}

		// Extract variables.
		if len(step.Extract) > 0 {
			extracted := make([]string, 0, len(step.Extract))
			for _, e := range step.Extract {
				val, err := evalJQ(e.JQ, jsonData)
				if err != nil {
					fmt.Printf("  WARNING: extract %q jq %q error: %v\n", e.Var, e.JQ, err)
					continue
				}
				strVal := fmt.Sprintf("%v", val)
				r.vars[e.Var] = strVal
				extracted = append(extracted, fmt.Sprintf("%s=%v", e.Var, val))
			}
			if len(extracted) > 0 {
				fmt.Printf("  \u2192 extracted: %s\n", strings.Join(extracted, ", "))
			}
		}
	}

	fmt.Printf("\nDone: %d steps, %d failures\n", total, failures)
	return failures, nil
}

// substitute replaces all {{var}} occurrences in s with known variable values.
func (r *Runner) substitute(s string) string {
	for k, v := range r.vars {
		s = strings.ReplaceAll(s, "{{"+k+"}}", v)
	}
	return s
}

// isAbsoluteURL returns true if the URL has a scheme (i.e. starts with http:// or https://).
func isAbsoluteURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	return err == nil && u.Scheme != ""
}

// evalJQ runs a jq expression against data and returns the first result value.
func evalJQ(expr string, data interface{}) (interface{}, error) {
	query, err := gojq.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("parse jq expression %q: %w", expr, err)
	}
	iter := query.Run(data)
	val, ok := iter.Next()
	if !ok {
		return nil, fmt.Errorf("jq expression %q produced no results", expr)
	}
	if err, ok := val.(error); ok {
		return nil, fmt.Errorf("jq expression %q: %w", expr, err)
	}
	return val, nil
}

// assertEquality compares a jq result value to an expected value from the YAML assert block.
// It handles numeric type coercion (YAML integers vs JSON float64).
func assertEquality(actual, expected interface{}) bool {
	if actual == expected {
		return true
	}
	// Normalise numbers: JSON numbers are float64; YAML booleans/ints may differ.
	actualStr := fmt.Sprintf("%v", actual)
	expectedStr := fmt.Sprintf("%v", expected)
	return actualStr == expectedStr
}
