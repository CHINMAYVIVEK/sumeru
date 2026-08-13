package web

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

const (
	setupRateLimitWindow = time.Minute
	setupRateLimitMax    = 5
)

var setupAttemptMu sync.Mutex
var setupAttempts = map[string][]time.Time{}

func clientIP(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		if i := strings.Index(xff, ","); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return strings.TrimSpace(r.RemoteAddr)
	}
	return host
}

func isLoopbackIP(ip string) bool {
	parsed := net.ParseIP(strings.TrimSpace(ip))
	return parsed != nil && parsed.IsLoopback()
}

// allowSetupRequest enforces localhost-only (when configured), optional setup token, and rate limits.
func allowSetupRequest(w http.ResponseWriter, r *http.Request, tokenFromBody string) bool {
	if orm.IsInitialized() {
		http.Error(w, "Setup already completed", http.StatusForbidden)
		return false
	}
	if config.AppConfig.SetupLocalhostOnly && !isLoopbackIP(clientIP(r)) {
		http.Error(w, "Setup is restricted to localhost", http.StatusForbidden)
		return false
	}
	expected := strings.TrimSpace(config.AppConfig.SetupToken)
	if expected != "" {
		got := strings.TrimSpace(r.Header.Get("X-Setup-Token"))
		if got == "" {
			got = strings.TrimSpace(tokenFromBody)
		}
		if got != expected {
			http.Error(w, "Invalid setup token", http.StatusForbidden)
			return false
		}
	}
	ip := clientIP(r)
	now := time.Now()
	setupAttemptMu.Lock()
	defer setupAttemptMu.Unlock()
	var kept []time.Time
	for _, t := range setupAttempts[ip] {
		if now.Sub(t) <= setupRateLimitWindow {
			kept = append(kept, t)
		}
	}
	if len(kept) >= setupRateLimitMax {
		http.Error(w, "Too many setup attempts", http.StatusTooManyRequests)
		setupAttempts[ip] = kept
		return false
	}
	setupAttempts[ip] = append(kept, now)
	return true
}
