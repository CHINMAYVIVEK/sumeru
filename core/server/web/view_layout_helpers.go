package web

import (
	"net/http"
	"strings"
)

// normalizeGridListLayout returns "grid" or "list". Maps legacy "kanban" to "grid".
func normalizeGridListLayout(raw string) string {
	layout := strings.ToLower(strings.TrimSpace(raw))
	if layout == "" || layout == "kanban" {
		return "grid"
	}
	if layout != "list" {
		return "grid"
	}
	return layout
}

// layoutFromQuery reads ?layout= from the request query string.
func layoutFromQuery(r *http.Request) string {
	return normalizeGridListLayout(r.URL.Query().Get("layout"))
}

// layoutFromForm reads a layout value from POST form fields (e.g. apps_layout).
func layoutFromForm(r *http.Request, field string) string {
	return normalizeGridListLayout(r.FormValue(field))
}
