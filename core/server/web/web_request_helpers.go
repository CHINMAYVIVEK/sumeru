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
func WebLogEvent(ctx context.Context, route, message, operation, status string, err error, contextFields map[string]interface{}) {
	if contextFields == nil {
		contextFields = map[string]interface{}{}
	}
	contextFields["route"] = route

	event := applog.Event{
		Message:   message,
		Component: webLogComponent,
		Operation: operation,
		Status:    status,
		Context:   contextFields,
		Err:       err,
	}
	emitWebLogEvent(ctx, event)
}

func emitWebLogEvent(ctx context.Context, event applog.Event) {
	switch {
	case event.Err != nil || event.Status == logStatusFailure:
		if event.Status == "" {
			event.Status = logStatusFailure
		}
		applog.Error(ctx, event)
	case event.Status == logStatusPartial:
		applog.Warn(ctx, event)
	default:
		applog.Info(ctx, event)
	}
}

func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	if route = strings.TrimSpace(route); route == "" {
		route = webLogUnknownRoute
	}
	WebLogEvent(ctx, route, fmt.Sprintf(format, args...), logOperationRequest, logStatusSuccess, nil, nil)
}

// WebLogNavigation emits an INFO-level audit event for successful navigation (menu, view, module, company).
func WebLogNavigation(ctx context.Context, route, operation, message string, fields map[string]interface{}) {
	WebLogEvent(ctx, route, message, operation, logStatusSuccess, nil, fields)
}

// SafePathNext returns a same-origin relative path, rejecting open-redirect targets.
func SafePathNext(rawURL, fallback string) string {
	return safeRedirectPath(rawURL, "/", fallback)
}

// SafeWebNext returns a path under /web, rejecting open-redirect targets.
func SafeWebNext(rawURL, fallback string) string {
	return safeRedirectPath(rawURL, "/web", fallback)
}

func safeRedirectPath(rawURL, requiredPrefix, fallback string) string {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" || !strings.HasPrefix(trimmed, requiredPrefix) || strings.HasPrefix(trimmed, "//") {
		return fallback
	}
	return trimmed
}

// redirectWithWebMessage redirects to a safe /web path with a flash message query parameter.
func redirectWithWebMessage(w http.ResponseWriter, r *http.Request, rawNext, message string) {
	nextPath := SafeWebNext(rawNext, homeRoute)
	redirectURL, err := urlWithQueryParam(nextPath, flashMessageParam, message)
	if err != nil {
		redirectURL = homeRoute + "?" + flashMessageParam + "=" + url.QueryEscape(message)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func urlWithQueryParam(path, param, value string) (string, error) {
	parsed, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	query := parsed.Query()
	query.Set(param, value)
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func RequirePOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	http.Error(w, methodNotAllowedMessage, http.StatusMethodNotAllowed)
	return false
}

func ParsePostForm(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, invalidFormMessage, http.StatusBadRequest)
		return false
	}
	return true
}

// requireLoginAndPOST checks login, POST method, form parsing, and CSRF (form body must be parsed first).
func requireLoginAndPOST(w http.ResponseWriter, r *http.Request) bool {
	return requireAuthenticatedPOST(w, r) &&
		ParsePostForm(w, r) &&
		validateSessionCSRF(w, r)
}

// requireLoginJSONPost checks login, POST method, and CSRF for JSON API handlers.
func requireLoginJSONPost(w http.ResponseWriter, r *http.Request) bool {
	return requireAuthenticatedPOST(w, r) && validateSessionCSRF(w, r)
}

// requireLoginMultipartPost checks login, POST method, bounded multipart parsing, and CSRF.
func requireLoginMultipartPost(w http.ResponseWriter, r *http.Request, maxBytes int64) bool {
	return requireAuthenticatedPOST(w, r) &&
		parseBoundedMultipartForm(w, r, maxBytes) &&
		validateSessionCSRF(w, r)
}

func requireAuthenticatedPOST(w http.ResponseWriter, r *http.Request) bool {
	return requireLogin(w, r) && RequirePOST(w, r)
}

func parseBoundedMultipartForm(w http.ResponseWriter, r *http.Request, maxBytes int64) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		http.Error(w, invalidFormMessage, http.StatusBadRequest)
		return false
	}
	return true
}

func validateSessionCSRF(w http.ResponseWriter, r *http.Request) bool {
	if ValidateCSRF(r) {
		return true
	}
	http.Error(w, invalidCSRFMessage, http.StatusForbidden)
	return false
}

func requireRegisteredModel(w http.ResponseWriter, modelName string) (orm.Model, bool) {
	model, registered := orm.Registry[modelName]
	if !registered || model == nil {
		http.Error(w, unknownModelMessage, http.StatusBadRequest)
		return nil, false
	}
	return model, true
}
