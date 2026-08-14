package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// WebLogEvent logs a structured web event using the applog contract.
func WebLogEvent(ctx context.Context, route, message, operation, status string, err error, ctxFields map[string]interface{}) {
	if ctxFields == nil {
		ctxFields = map[string]interface{}{}
	}
	ctxFields["route"] = route
	ev := applog.Event{
		Message:   message,
		Component: "web",
		Operation: operation,
		Status:    status,
		Context:   ctxFields,
		Err:       err,
	}
	if err != nil || status == "failure" {
		if ev.Status == "" {
			ev.Status = "failure"
		}
		applog.Error(ctx, ev)
		return
	}
	if status == "partial" {
		applog.Warn(ctx, ev)
		return
	}
	applog.Info(ctx, ev)
}

func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	WebLogEvent(ctx, route, fmt.Sprintf(format, args...), "request", "success", nil, nil)
}

// WebLogNavigation emits an INFO-level audit event for successful navigation (menu, view, module, company).
func WebLogNavigation(ctx context.Context, route, operation, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = map[string]interface{}{}
	}
	WebLogEvent(ctx, route, message, operation, "success", nil, fields)
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

// redirectWithWebMessage redirects to a safe /web path with a msg query parameter.
func redirectWithWebMessage(w http.ResponseWriter, r *http.Request, rawNext, message string) {
	next := SafeWebNext(rawNext, homeRoute)
	parsed, err := url.Parse(next)
	if err != nil {
		http.Redirect(w, r, homeRoute+"?msg="+url.QueryEscape(message), http.StatusSeeOther)
		return
	}
	query := parsed.Query()
	query.Set("msg", message)
	parsed.RawQuery = query.Encode()
	http.Redirect(w, r, parsed.String(), http.StatusSeeOther)
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

// requireLoginAndPOST checks login, method, CSRF, and parses the form.
func requireLoginAndPOST(w http.ResponseWriter, r *http.Request) bool {
	if !requireLogin(w, r) || !RequirePOST(w, r) || !ParsePostForm(w, r) {
		return false
	}
	if !ValidateCSRF(r) {
		http.Error(w, invalidCSRFMessage, http.StatusForbidden)
		return false
	}
	return true
}

// requireLoginJSONPost checks login, POST method, and CSRF for JSON API handlers.
func requireLoginJSONPost(w http.ResponseWriter, r *http.Request) bool {
	if !requireLogin(w, r) || !RequirePOST(w, r) {
		return false
	}
	if !ValidateCSRF(r) {
		http.Error(w, invalidCSRFMessage, http.StatusForbidden)
		return false
	}
	return true
}

// requireLoginMultipartPost checks login, POST method, CSRF, and parses a bounded multipart form.
func requireLoginMultipartPost(w http.ResponseWriter, r *http.Request, maxBytes int64) bool {
	if !requireLogin(w, r) || !RequirePOST(w, r) {
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return false
	}
	if !ValidateCSRF(r) {
		http.Error(w, invalidCSRFMessage, http.StatusForbidden)
		return false
	}
	return true
}

func validateSessionCSRF(w http.ResponseWriter, r *http.Request) bool {
	if !ValidateCSRF(r) {
		http.Error(w, invalidCSRFMessage, http.StatusForbidden)
		return false
	}
	return true
}

func requireRegisteredModel(w http.ResponseWriter, modelName string) (orm.Model, bool) {
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		http.Error(w, "Unknown model", http.StatusBadRequest)
		return nil, false
	}
	return inst, true
}
