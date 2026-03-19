package auth

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/smilemilks2021/easy-web/internal/browser"
	"github.com/smilemilks2021/easy-web/internal/config"
	"github.com/smilemilks2021/easy-web/internal/cookie"
)

type Result struct {
	Cookies    []*cookie.Entry
	TokenValue string
}

type Options struct {
	Mode                string
	UseEmbeddedChromium bool
	NoReuseProfile      bool
	VerboseAuth         bool
	CaptureHeader       string
	ExtractToken        bool
	ChromiumExecPath    string
}

func Resolve(targetURL string, opts Options) (*Result, error) {
	domain := parseHost(targetURL)
	store := cookie.NewCache(config.CacheDir())

	switch opts.Mode {
	case "auto":
		return autoFallback(targetURL, domain, store, opts)
	case "chrome":
		return fromChrome(domain, store)
	case "chromedp":
		return fromChromedp(targetURL, domain, store, opts, true)
	case "browser":
		return fromBrowser(targetURL, domain, store, opts)
	case "remote":
		return fromRemote(targetURL, domain, store, opts)
	default:
		return nil, fmt.Errorf("unknown mode: %s", opts.Mode)
	}
}

func autoFallback(targetURL, domain string, store *cookie.Cache, opts Options) (*Result, error) {
	// Step 1: local cache — skip if any JWT cookie is expired
	if entries, err := store.Load(domain); err == nil && len(entries) > 0 {
		expired := false
		for _, e := range entries {
			if isJWTCookie(e.Name) && IsJWTExpired(e.Value) {
				expired = true
				break
			}
		}
		if !expired {
			if opts.VerboseAuth {
				fmt.Printf("[auth] cache hit for %s\n", domain)
			}
			return &Result{Cookies: entries}, nil
		}
		if opts.VerboseAuth {
			fmt.Printf("[auth] cached JWT expired for %s, re-auth\n", domain)
		}
		_ = store.Delete(domain)
	}

	// Step 2: system Chrome
	if !opts.UseEmbeddedChromium && chromeInstalled() {
		if result, err := fromChrome(domain, store); err == nil {
			return result, nil
		} else if opts.VerboseAuth {
			fmt.Printf("[auth] Chrome DB failed: %v\n", err)
		}
	}

	// Step 3: chromedp
	return fromChromedp(targetURL, domain, store, opts, true)
}

func isJWTCookie(name string) bool {
	for _, n := range config.C.Domains.JWTCookies {
		if n == name {
			return true
		}
	}
	return false
}

func fromChrome(domain string, store *cookie.Cache) (*Result, error) {
	entries, err := cookie.ReadChromeCookies(domain)
	if err != nil {
		return nil, err
	}
	_ = store.Save(domain, entries)
	return &Result{Cookies: entries}, nil
}

func fromChromedp(targetURL, domain string, store *cookie.Cache, opts Options, headless bool) (*Result, error) {
	d := browser.NewDriver(browser.Options{
		Headless:     headless,
		ReuseProfile: !opts.NoReuseProfile,
		ProfileDir:   config.ProfileDir(),
		ExecPath:     opts.ChromiumExecPath,
		VerboseAuth:  opts.VerboseAuth,
	})
	entries, err := d.LoginAndGetCookies(targetURL, 5*time.Minute)
	if err != nil {
		return nil, err
	}
	_ = store.Save(domain, entries)
	return &Result{Cookies: entries}, nil
}

func fromBrowser(targetURL, domain string, store *cookie.Cache, opts Options) (*Result, error) {
	d := browser.NewDriver(browser.Options{
		Headless:     false,
		ReuseProfile: !opts.NoReuseProfile,
		ProfileDir:   config.ProfileDir(),
		ExecPath:     opts.ChromiumExecPath,
	})
	entries, err := d.LoginAndGetCookies(targetURL, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	_ = store.Save(domain, entries)
	return &Result{Cookies: entries}, nil
}

func fromRemote(targetURL, domain string, store *cookie.Cache, opts Options) (*Result, error) {
	debugPort := config.C.DebugPort
	if debugPort == 0 {
		debugPort = 9222
	}
	d := browser.NewDriver(browser.Options{})
	entries, err := d.LoginRemote(targetURL, debugPort, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("remote mode (port %d): %w", debugPort, err)
	}
	_ = store.Save(domain, entries)
	return &Result{Cookies: entries}, nil
}

func chromeInstalled() bool {
	for _, p := range chromePaths() {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func chromePaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}
	case "linux":
		home, _ := os.UserHomeDir()
		return []string{
			"/usr/bin/google-chrome",
			"/usr/bin/chromium-browser",
			filepath.Join(home, ".config/google-chrome/Default/Cookies"),
		}
	case "windows":
		la, pf := os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES")
		return []string{
			filepath.Join(la, "Google", "Chrome", "Application", "chrome.exe"),
			filepath.Join(pf, "Google", "Chrome", "Application", "chrome.exe"),
		}
	}
	return nil
}

func parseHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Host == "" {
		return rawURL
	}
	return u.Hostname()
}
