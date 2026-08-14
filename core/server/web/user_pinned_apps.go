package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
)

type pinnedAppsResponse struct {
	Modules []string `json:"modules"`
}

// PinnedAppsSaveHandler persists the user's pinned application modules.
func PinnedAppsSaveHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginJSONPost(w, r) {
		return
	}

	modules, ok := parsePinnedAppsRequest(r)
	if !ok {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clean, err := render.SavePinnedAppsForUser(r.Context(), modules)
	if err != nil {
		WebLogEvent(r.Context(), r.URL.Path, "Failed to save pinned apps", "save", "failure", err, nil)
		http.Error(w, "Failed to save pinned apps", http.StatusInternalServerError)
		return
	}
	if clean == nil {
		clean = []string{}
	}

	writeJSON(w, r.Context(), r.URL.Path, pinnedAppsResponse{Modules: clean})
}

func parsePinnedAppsRequest(r *http.Request) ([]string, bool) {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return nil, false
		}
		var payload struct {
			Modules []string `json:"modules"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, false
		}
		return payload.Modules, true
	}
	if err := r.ParseForm(); err != nil {
		return nil, false
	}
	raw := strings.TrimSpace(r.FormValue("modules"))
	if raw == "" {
		return []string{}, true
	}
	var modules []string
	if err := json.Unmarshal([]byte(raw), &modules); err != nil {
		return nil, false
	}
	return modules, true
}
