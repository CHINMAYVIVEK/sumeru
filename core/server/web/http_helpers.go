package web

import (
	"net/url"
	"strconv"
	"strings"
)

func splitViewModes(viewMode string) []string {
	var out []string
	for _, p := range splitComma(viewMode) {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitComma(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func normalizeViewMode(mode string) string {
	return strings.ToLower(strings.TrimSpace(mode))
}

// formBaseQueryValues builds a stable /web query string (no leading "?") for form Edit/Cancel/Save redirects.
func formBaseQueryValues(actionID int, menuID, viewType, idStr string) string {
	q := url.Values{}
	if actionID > 0 {
		q.Set("action", strconv.Itoa(actionID))
	}
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	if vt := strings.TrimSpace(viewType); vt != "" {
		q.Set("view_type", vt)
	}
	if strings.TrimSpace(idStr) != "" {
		q.Set("id", strings.TrimSpace(idStr))
	}
	return q.Encode()
}
