package request_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/smilemilks2021/easy-web/internal/cookie"
	"github.com/smilemilks2021/easy-web/internal/request"
)

func TestClientSendsCookies(t *testing.T) {
	var gotCookie string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCookie = r.Header.Get("Cookie")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c := request.NewClient([]*cookie.Entry{
		{Name: "session", Value: "tok123", Domain: "127.0.0.1", Expires: time.Now().Add(time.Hour)},
	}, nil)
	if _, err := c.Do("GET", ts.URL, "", nil); err != nil {
		t.Fatal(err)
	}
	if !contains(gotCookie, "session=tok123") {
		t.Fatalf("expected cookie header, got: %s", gotCookie)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}
func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
