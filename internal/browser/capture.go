package browser

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type CapturedRequest struct {
	URL     string
	Method  string
	Headers map[string]string
}

type CaptureOptions struct {
	Patterns     []string      // URL substrings to include (OR); empty = all
	Timeout      time.Duration
	ExecPath     string
	ProfileDir   string
	ReuseProfile bool
}

func CaptureRequests(targetURL string, opts CaptureOptions) ([]*CapturedRequest, error) {
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun, chromedp.NoDefaultBrowserCheck,
	}
	if opts.ExecPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(opts.ExecPath))
	}
	if opts.ReuseProfile && opts.ProfileDir != "" {
		allocOpts = append(allocOpts, chromedp.UserDataDir(opts.ProfileDir))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTO := context.WithTimeout(ctx, opts.Timeout)
	defer cancelTO()

	var (
		mu       sync.Mutex
		captured []*CapturedRequest
	)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		req, ok := ev.(*network.EventRequestWillBeSent)
		if !ok {
			return
		}
		if !matchesPatterns(req.Request.URL, opts.Patterns) {
			return
		}
		headers := make(map[string]string)
		for k, v := range req.Request.Headers {
			if s, ok := v.(string); ok {
				headers[k] = s
			}
		}
		mu.Lock()
		captured = append(captured, &CapturedRequest{
			URL:     req.Request.URL,
			Method:  req.Request.Method,
			Headers: headers,
		})
		mu.Unlock()
	})

	if err := chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(targetURL),
	); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	fmt.Printf("Recording API requests (timeout: %s). Press Enter to stop.\n", opts.Timeout)
	done := make(chan struct{}, 1)
	go func() { fmt.Scanln(); close(done) }()
	select {
	case <-done:
	case <-ctx.Done():
		fmt.Println("Capture timeout reached.")
	}
	mu.Lock()
	result := captured
	mu.Unlock()
	return result, nil
}

func matchesPatterns(url string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if strings.Contains(url, p) {
			return true
		}
	}
	return false
}
