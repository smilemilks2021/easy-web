package request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/smilemilks2021/easy-web/internal/cookie"
)

// Options configures optional behaviour for Client.
type Options struct {
	Retry      int
	RetryDelay time.Duration
	Timeout    time.Duration // defaults to 30s when zero
	ProxyURL   string
}

type Client struct {
	cookies      []*cookie.Entry
	extraHeaders map[string]string
	opts         Options
	http         *http.Client
}

// NewClient creates a Client. Pass a zero-value Options{} for default behaviour.
// The old two-argument signature is preserved for callers that don't need Options:
// existing call sites that pass (cookies, extraHeaders) continue to compile because
// Go does not support overloading – callers must now pass Options explicitly.
// To keep the existing test working, NewClient accepts options as a variadic arg.
func NewClient(cookies []*cookie.Entry, extraHeaders map[string]string, opts ...Options) (*Client, error) {
	o := Options{}
	if len(opts) > 0 {
		o = opts[0]
	}
	if o.Timeout == 0 {
		o.Timeout = 30 * time.Second
	}
	if o.RetryDelay == 0 {
		o.RetryDelay = time.Second
	}

	// #3 fix: safe type assertion — don't panic if DefaultTransport was replaced
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok || base == nil {
		base = &http.Transport{}
	}
	transport := base.Clone()

	// #7 fix: return error on invalid proxy URL instead of silently ignoring
	if o.ProxyURL != "" {
		pu, parseErr := url.Parse(o.ProxyURL)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", o.ProxyURL, parseErr)
		}
		transport.Proxy = http.ProxyURL(pu)
	}

	return &Client{
		cookies:      cookies,
		extraHeaders: extraHeaders,
		opts:         o,
		http:         &http.Client{Timeout: o.Timeout, Transport: transport},
	}, nil
}

// Do performs the HTTP request with retry logic applied.
func (c *Client) Do(method, rawURL, body string, callHeaders map[string]string) (*http.Response, error) {
	maxAttempts := c.opts.Retry + 1
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			fmt.Printf("[retry %d/%d] %s %s\n", attempt, c.opts.Retry, method, rawURL)
			time.Sleep(c.opts.RetryDelay)
		}
		resp, err = c.doOnce(method, rawURL, body, callHeaders)
		if err == nil {
			if resp.StatusCode < 500 {
				return resp, nil
			}
		}
		// If we have a response body on a non-retryable or final attempt, drain it.
		if resp != nil && attempt < maxAttempts-1 {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
	return resp, err
}

func (c *Client) doOnce(method, rawURL, body string, callHeaders map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, rawURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if len(c.cookies) > 0 {
		req.Header.Set("Cookie", cookie.CookieHeader(c.cookies))
	}
	for k, v := range c.extraHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range callHeaders {
		req.Header.Set(k, v)
	}
	if body != "" {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}
	return c.http.Do(req)
}
