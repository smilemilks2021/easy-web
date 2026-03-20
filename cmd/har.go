package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/smilemilks2021/easy-web/internal/browser"
)

// HAR 1.2 types (core fields only).

type harFile struct {
	Log harLog `json:"log"`
}

type harLog struct {
	Version string     `json:"version"`
	Creator harCreator `json:"creator"`
	Entries []harEntry `json:"entries"`
}

type harCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type harEntry struct {
	StartedDateTime string     `json:"startedDateTime"`
	Time            float64    `json:"time"`
	Request         harRequest `json:"request"`
	Response        harResponse `json:"response"`
	Cache           struct{}   `json:"cache"`
	Timings         harTimings `json:"timings"`
}

type harRequest struct {
	Method      string      `json:"method"`
	URL         string      `json:"url"`
	HTTPVersion string      `json:"httpVersion"`
	Headers     []harNVP    `json:"headers"`
	QueryString []harNVP    `json:"queryString"`
	Cookies     []harNVP    `json:"cookies"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
	PostData    *harPostData `json:"postData,omitempty"`
}

type harResponse struct {
	Status      int      `json:"status"`
	StatusText  string   `json:"statusText"`
	HTTPVersion string   `json:"httpVersion"`
	Headers     []harNVP `json:"headers"`
	Cookies     []harNVP `json:"cookies"`
	Content     harContent `json:"content"`
	RedirectURL string   `json:"redirectURL"`
	HeadersSize int      `json:"headersSize"`
	BodySize    int      `json:"bodySize"`
}

type harContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
}

type harPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type harNVP struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harTimings struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

// exportHAR writes captured requests to outputPath in HAR 1.2 JSON format.
// baseURL is used for context only (e.g. to normalise relative URLs if needed).
func exportHAR(reqs []*browser.CapturedRequest, baseURL, outputPath string) error {
	entries := make([]harEntry, 0, len(reqs))
	now := time.Now().UTC().Format(time.RFC3339)

	for _, r := range reqs {
		// Build query string pairs.
		var queryString []harNVP
		if u, err := url.Parse(r.URL); err == nil {
			for k, vs := range u.Query() {
				for _, v := range vs {
					queryString = append(queryString, harNVP{Name: k, Value: v})
				}
			}
		}

		// Build request headers.
		var headers []harNVP
		for k, v := range r.Headers {
			headers = append(headers, harNVP{Name: k, Value: v})
		}

		req := harRequest{
			Method:      r.Method,
			URL:         r.URL,
			HTTPVersion: "HTTP/1.1",
			Headers:     headers,
			QueryString: queryString,
			Cookies:     []harNVP{},
			HeadersSize: -1,
			BodySize:    -1,
		}
		if r.Body != "" {
			mimeType := "application/json"
			if ct, ok := r.Headers["Content-Type"]; ok {
				mimeType = ct
			}
			req.PostData = &harPostData{
				MimeType: mimeType,
				Text:     r.Body,
			}
			req.BodySize = len(r.Body)
		}

		entry := harEntry{
			StartedDateTime: now,
			Time:            0,
			Request:         req,
			Response: harResponse{
				Status:      0,
				StatusText:  "",
				HTTPVersion: "HTTP/1.1",
				Headers:     []harNVP{},
				Cookies:     []harNVP{},
				Content:     harContent{Size: 0, MimeType: ""},
				RedirectURL: "",
				HeadersSize: -1,
				BodySize:    -1,
			},
			Timings: harTimings{Send: 0, Wait: 0, Receive: 0},
		}
		entries = append(entries, entry)
	}

	version := appVersion
	if version == "" {
		version = "dev"
	}
	har := harFile{
		Log: harLog{
			Version: "1.2",
			Creator: harCreator{Name: "easy-web", Version: version},
			Entries: entries,
		},
	}

	data, err := json.MarshalIndent(har, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal HAR: %w", err)
	}
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("write HAR file: %w", err)
	}
	fmt.Printf("Exported %d requests to %s (HAR 1.2)\n", len(reqs), outputPath)
	return nil
}
