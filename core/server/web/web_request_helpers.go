package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	msg := fmt.Sprintf(format, args...)
	applog.L(ctx).Info("web", "route", route, "msg", msg)
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

// requireLoginAndPOST checks login, method, and parses the form.
func requireLoginAndPOST(w http.ResponseWriter, r *http.Request) bool {
	return requireLogin(w, r) && RequirePOST(w, r) && ParsePostForm(w, r)
}

func requireRegisteredModel(w http.ResponseWriter, modelName string) (orm.Model, bool) {
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		http.Error(w, "Unknown model", http.StatusBadRequest)
		return nil, false
	}
	return inst, true
}
