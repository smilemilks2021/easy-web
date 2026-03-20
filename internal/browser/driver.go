package browser

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/smilemilks2021/easy-web/internal/cookie"
)

type Options struct {
	Headless        bool
	ReuseProfile    bool
	ProfileDir      string
	ExecPath        string
	RemoteDebugPort int
	VerboseAuth     bool
}

type Driver struct{ opts Options }

func NewDriver(opts Options) *Driver { return &Driver{opts: opts} }

func DefaultProfileDir() string {
	home, _ := os.UserHomeDir()
	return home + "/.easy-web/chrome-data"
}

// expiresTime converts a float64 Unix seconds value (from network.Cookie.Expires)
// to time.Time. A value <= 0 or not-finite is treated as zero (session cookie).
func expiresTime(f float64) time.Time {
	if f <= 0 || math.IsInf(f, 0) || math.IsNaN(f) {
		return time.Time{}
	}
	sec := int64(f)
	nsec := int64((f - float64(sec)) * 1e9)
	return time.Unix(sec, nsec)
}

// LoginAndGetCookies navigates to targetURL, waits for login, returns all cookies.
func (d *Driver) LoginAndGetCookies(targetURL string, timeout time.Duration) ([]*cookie.Entry, error) {
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}
	if d.opts.Headless {
		allocOpts = append(allocOpts, chromedp.Headless, chromedp.DisableGPU)
	}
	if d.opts.ExecPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(d.opts.ExecPath))
	}
	if d.opts.ReuseProfile && d.opts.ProfileDir != "" {
		os.MkdirAll(d.opts.ProfileDir, 0700)
		allocOpts = append(allocOpts, chromedp.UserDataDir(d.opts.ProfileDir))
	}

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTO := context.WithTimeout(ctx, timeout)
	defer cancelTO()

	if err := chromedp.Run(ctx, chromedp.Navigate(targetURL)); err != nil {
		return nil, fmt.Errorf("navigate: %w", err)
	}

	if !d.opts.Headless {
		fmt.Println("Browser opened. Login, then press Enter to capture cookies...")
		fmt.Scanln()
	} else {
		time.Sleep(2 * time.Second)
	}

	// Use cdproto/network types (not chromedp.Cookie which doesn't exist)
	var rawCookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		rawCookies, err = network.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return nil, fmt.Errorf("get cookies: %w", err)
	}

	entries := make([]*cookie.Entry, 0, len(rawCookies))
	for _, c := range rawCookies {
		entries = append(entries, &cookie.Entry{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			Expires:  expiresTime(c.Expires), // float64 Unix seconds → time.Time
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
		})
	}
	return entries, nil
}

// LoginRemote connects to an already-running Chrome via CDP WebSocket.
func (d *Driver) LoginRemote(targetURL string, debugPort int, timeout time.Duration) ([]*cookie.Entry, error) {
	wsURL := fmt.Sprintf("http://127.0.0.1:%d", debugPort)
	allocCtx, cancelAlloc := chromedp.NewRemoteAllocator(context.Background(), wsURL)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTO := context.WithTimeout(ctx, timeout)
	defer cancelTO()

	if err := chromedp.Run(ctx, chromedp.Navigate(targetURL)); err != nil {
		return nil, fmt.Errorf("navigate via remote: %w", err)
	}
	time.Sleep(time.Second)

	var rawCookies []*network.Cookie
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		rawCookies, err = network.GetCookies().Do(ctx)
		return err
	})); err != nil {
		return nil, err
	}

	entries := make([]*cookie.Entry, 0, len(rawCookies))
	for _, c := range rawCookies {
		entries = append(entries, &cookie.Entry{
			Name: c.Name, Value: c.Value, Domain: c.Domain,
			Path: c.Path, Expires: expiresTime(c.Expires),
			Secure: c.Secure, HTTPOnly: c.HTTPOnly,
		})
	}
	return entries, nil
}

// ExtractLocalStorageToken returns the first matching token from localStorage/sessionStorage.
func (d *Driver) ExtractLocalStorageToken(targetURL string, keys []string) (string, error) {
	allocOpts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun, chromedp.Headless, chromedp.DisableGPU,
	}
	if d.opts.ExecPath != "" {
		allocOpts = append(allocOpts, chromedp.ExecPath(d.opts.ExecPath))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
	defer cancelAlloc()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancelTO := context.WithTimeout(ctx, 30*time.Second)
	defer cancelTO()

	if err := chromedp.Run(ctx, chromedp.Navigate(targetURL)); err != nil {
		return "", err
	}
	time.Sleep(2 * time.Second)

	for _, key := range keys {
		var val string
		script := fmt.Sprintf(`localStorage.getItem(%q) || sessionStorage.getItem(%q) || ""`, key, key)
		if err := chromedp.Run(ctx, chromedp.Evaluate(script, &val)); err == nil && val != "" {
			return val, nil
		}
	}
	return "", fmt.Errorf("token not found for keys: %v", keys)
}
