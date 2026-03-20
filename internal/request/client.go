package request

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/smilemilks2021/easy-web/internal/cookie"
)

type Client struct {
	cookies      []*cookie.Entry
	extraHeaders map[string]string
	http         *http.Client
}

func NewClient(cookies []*cookie.Entry, extraHeaders map[string]string) *Client {
	return &Client{cookies: cookies, extraHeaders: extraHeaders, http: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) Do(method, url, body string, callHeaders map[string]string) (*http.Response, error) {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = bytes.NewBufferString(body)
	}
	req, err := http.NewRequest(method, url, bodyReader)
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
		req.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(req)
}
