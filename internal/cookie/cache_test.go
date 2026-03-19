package cookie_test

import (
	"testing"
	"time"

	"github.com/smilemilks2021/easy-web/internal/cookie"
)

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := cookie.NewCache(dir)
	entries := []*cookie.Entry{{Name: "s", Value: "abc", Domain: "ex.com", Expires: time.Now().Add(time.Hour)}}
	if err := c.Save("ex.com", entries); err != nil {
		t.Fatal(err)
	}
	got, err := c.Load("ex.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "abc" {
		t.Fatalf("got %v", got)
	}
}

func TestCacheList(t *testing.T) {
	dir := t.TempDir()
	c := cookie.NewCache(dir)
	_ = c.Save("a.com", []*cookie.Entry{{Name: "x", Value: "1", Domain: "a.com"}})
	domains, err := c.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(domains) != 1 {
		t.Fatalf("expected 1 domain, got %d", len(domains))
	}
}
