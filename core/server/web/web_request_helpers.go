package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/applog"
)

func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	msg := fmt.Sprintf(format, args...)
	applog.L(ctx).Infow("web", "route", route, "msg", msg)
}

func SafePathNext(raw, fallback string) string {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return fallback
	}
	return s
}

func SafeWebNext(raw, fallback string) string {
	s := strings.TrimSpace(raw)
	if s == "" || !strings.HasPrefix(s, "/web") || strings.HasPrefix(s, "//") {
		return fallback
	}
	return s
}

func RequirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return false
}

func ParsePostForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return false
	}
	return true
}
