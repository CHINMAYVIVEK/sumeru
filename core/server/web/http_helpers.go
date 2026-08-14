package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func splitViewModes(viewMode string) []string {
	return splitCommaSeparatedValues(viewMode)
}

func splitCommaSeparatedValues(raw string) []string {
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

func normalizeViewMode(viewMode string) string {
	return strings.ToLower(strings.TrimSpace(viewMode))
}

// formBaseQueryValues builds a stable /web query string (no leading "?") for form Edit/Cancel/Save redirects.
func formBaseQueryValues(actionID int, menuID, viewType, recordID string) string {
	query := url.Values{}
	setWorkspaceQueryInt(query, workspaceActionParam, actionID)
	setWorkspaceQueryString(query, workspaceMenuIDParam, menuID)
	setWorkspaceQueryString(query, workspaceViewTypeParam, viewType)
	setWorkspaceQueryString(query, workspaceRecordIDParam, recordID)
	return query.Encode()
}

func setWorkspaceQueryString(query url.Values, param, value string) {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		query.Set(param, trimmed)
	}
}

func setWorkspaceQueryInt(query url.Values, param string, value int) {
	if value > 0 {
		query.Set(param, strconv.Itoa(value))
	}
}

func formOrQueryValue(r *http.Request, field string) string {
	if value := strings.TrimSpace(r.FormValue(field)); value != "" {
		return value
	}
	return strings.TrimSpace(r.URL.Query().Get(field))
}

// workspaceListURL builds a /web list redirect after delete (action + menu_id only).
func workspaceListURL(actionID, menuID string) string {
	query := url.Values{}
	setWorkspaceQueryString(query, workspaceActionParam, actionID)
	setWorkspaceQueryString(query, workspaceMenuIDParam, menuID)
	encoded := query.Encode()
	if encoded == "" {
		return workspaceRoute
	}
	return workspaceRoute + "?" + encoded
}
