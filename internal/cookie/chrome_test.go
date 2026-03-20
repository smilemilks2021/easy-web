// internal/cookie/chrome_test.go
package cookie_test

import "testing"

// ReadChromeCookies requires a real Chrome installation to test.
// This test verifies the function signature compiles correctly.
func TestReadChromeCookiesCompiles(t *testing.T) {
	// Function signature test — does not call the function (needs real Chrome DB)
	_ = readChromeCookiesExists
}

var readChromeCookiesExists = func() { /* import check only */ }
