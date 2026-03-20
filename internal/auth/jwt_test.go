package auth_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/smilemilks2021/easy-web/internal/auth"
)

func jwt(exp int64) string {
	h := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	p, _ := json.Marshal(map[string]int64{"exp": exp})
	return fmt.Sprintf("%s.%s.sig", h, base64.RawURLEncoding.EncodeToString(p))
}

func TestJWTExpired(t *testing.T) {
	if !auth.IsJWTExpired(jwt(time.Now().Add(-time.Hour).Unix())) {
		t.Fatal("expected expired")
	}
}
func TestJWTValid(t *testing.T) {
	if auth.IsJWTExpired(jwt(time.Now().Add(time.Hour).Unix())) {
		t.Fatal("expected valid")
	}
}
func TestJWTInvalid(t *testing.T) {
	if auth.IsJWTExpired("not-a-jwt") {
		t.Fatal("invalid should not be expired")
	}
}
