package web

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
)

var (
	csrfSecretMu sync.RWMutex
	csrfSecret   []byte
)

func csrfKey() []byte {
	csrfSecretMu.RLock()
	if len(csrfSecret) > 0 {
		defer csrfSecretMu.RUnlock()
		return csrfSecret
	}
	csrfSecretMu.RUnlock()

	csrfSecretMu.Lock()
	defer csrfSecretMu.Unlock()
	if len(csrfSecret) == 0 {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			csrfSecret = []byte("sumeru-dev-csrf-fallback")
		} else {
			csrfSecret = b
		}
	}
	return csrfSecret
}

func sessionIDFromRequest(r *http.Request) string {
	c, err := r.Cookie(sessionCookieName)
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}

// CSRFTokenForRequest returns the per-session CSRF token (empty when not logged in).
func CSRFTokenForRequest(r *http.Request) string {
	sid := sessionIDFromRequest(r)
	if sid == "" {
		return ""
	}
	mac := hmac.New(sha256.New, csrfKey())
	_, _ = mac.Write([]byte(sid))
	return hex.EncodeToString(mac.Sum(nil)[:16])
}

// ValidateCSRF checks the csrf_token form field against the session-bound token.
func ValidateCSRF(r *http.Request) bool {
	expected := CSRFTokenForRequest(r)
	if expected == "" {
		return false
	}
	got := r.PostFormValue("csrf_token")
	if got == "" {
		got = r.Header.Get("X-CSRF-Token")
	}
	return got != "" && hmac.Equal([]byte(got), []byte(expected))
}
