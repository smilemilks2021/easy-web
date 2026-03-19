// internal/browser/chromium/downloader_test.go
package chromium_test

import (
	"testing"

	"github.com/smilemilks2021/easy-web/internal/browser/chromium"
)

func TestDefaultRevisionNotEmpty(t *testing.T) {
	if chromium.DefaultRevisionForTest() == "" {
		t.Fatal("DefaultRevision must not be empty")
	}
}
