package cookie

import (
	"context"
	"fmt"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/chrome"
)

// ReadChromeCookies reads and decrypts Chrome cookies for a domain using kooky.
// kooky handles DPAPI (Windows), Keychain (macOS), and libsecret/peanuts (Linux).
func ReadChromeCookies(domain string) ([]*Entry, error) {
	ctx := context.Background()
	stores := kooky.FindAllCookieStores(ctx)

	var result []*Entry
	for _, store := range stores {
		if !strings.EqualFold(store.Browser(), "chrome") {
			continue
		}
		seq := store.TraverseCookies(kooky.Domain(domain))
		for c, err := range seq {
			if err != nil {
				continue
			}
			if c == nil {
				continue
			}
			result = append(result, &Entry{
				Name:     c.Name,
				Value:    c.Value,
				Domain:   c.Domain,
				Path:     c.Path,
				Expires:  c.Expires,
				Secure:   c.Secure,
				HTTPOnly: c.HttpOnly,
			})
		}
		store.Close()
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no Chrome cookies for domain %q (Chrome may be running — close it and retry)", domain)
	}
	return result, nil
}
