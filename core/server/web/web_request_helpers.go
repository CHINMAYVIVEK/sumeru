package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/applog"
)

// WebLogf writes a structured web log line with enforced fields from applog.L(ctx) (user_id, log_ts, log_tz).
func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	msg := fmt.Sprintf(format, args...)
	applog.L(ctx).Infow("web", "route", route, "msg", msg)
}

// SafePathNext returns next if it is a same-site relative path (starts with /, not // open-redirect).
func SafePathNext(raw, fallback string) string {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return fallback
	}
	return s
}

// SafeWebNext returns next if it is under /web (and not an open-redirect via //).
func SafeWebNext(raw, fallback string) string {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.HasPrefix(s, "/web") || strings.HasPrefix(s, "//") {
		return fallback
	}
	return s
}

// RequirePOST writes 405 unless r.Method is POST.
func RequirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}

// ParsePostForm parses the body as application/x-www-form-urlencoded; on error writes 400 and returns false.
func ParsePostForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return false
	}
	return true
}
