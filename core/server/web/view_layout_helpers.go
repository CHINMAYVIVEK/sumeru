package web

import (
	"net/http"
	"strings"
)

// normalizeGridListLayout returns grid or list; empty, kanban, and unknown values default to grid.
func normalizeGridListLayout(raw string) string {
	normalizedLayout := strings.ToLower(strings.TrimSpace(raw))
	if normalizedLayout == appsLayoutList {
		return appsLayoutList
	}
	return appsLayoutGrid
}

func layoutFromQuery(r *http.Request) string {
	return normalizeGridListLayout(r.URL.Query().Get(layoutQueryParam))
}

func layoutFromForm(r *http.Request, fieldName string) string {
	return normalizeGridListLayout(r.FormValue(fieldName))
}
