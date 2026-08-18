package web

import (
	"encoding/json"
	"net/http"
)

const swcWorkspaceRoute = "/web/swc/workspace"

func registerSwcRoutes() {
	registerSession(http.MethodGet, swcWorkspaceRoute, SwcWorkspaceHandler)
	registerSwcBusRoute()
	registerSwcChatterRoute()
}

// SwcWorkspaceHandler GET /web/swc/workspace — JSON workspace payload for SWC.
func SwcWorkspaceHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	actionQuery, menuQuery := workspaceQueryParams(r)
	if redirectIfMenuAccessDenied(w, r, menuQuery) {
		return
	}
	actionID := ResolveWindowActionID(ctx, actionQuery, menuQuery)
	if actionID == 0 {
		http.Error(w, "action required", http.StatusBadRequest)
		return
	}
	actionData, err := loadWindowAction(ctx, actionID)
	if err != nil {
		respondActionNotFound(w, actionID)
		return
	}
	resolved, err := resolveWorkspaceView(ctx, r, actionData)
	if err != nil {
		http.Error(w, err.Error(), httpStatusFromWorkspaceError(err))
		return
	}
	req := parseWorkspaceRequest(r, actionID)
	viewRecord, err := buildViewRecordData(ctx, w, r, req, resolved, actionData)
	if err != nil {
		respondWorkspaceLoadError(w, ctx, err)
		return
	}
	payload := buildSwcWorkspacePayload(ctx, resolved, req, viewRecord)
	writeJSONResponse(w, payload)
}

func writeJSONResponse(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
