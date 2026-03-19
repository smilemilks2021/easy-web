// internal/browser/chromium/platform_test.go
package chromium_test

import (
	"testing"
	"github.com/smilemilks2021/easy-web/internal/browser/chromium"
)

func TestPlatformNotEmpty(t *testing.T) {
	p := chromium.Platform()
	if p == "" {
		t.Fatal("Platform() must not be empty")
	}
	allowed := map[string]bool{"mac-arm64": true, "mac-x64": true, "linux64": true, "win64": true}
	if !allowed[p] {
		t.Fatalf("unexpected platform: %s", p)
	}
}
