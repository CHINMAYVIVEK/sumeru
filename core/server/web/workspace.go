package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// WebHandler renders sys.action.window targets using sys.view definitions.
func WebHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	actionIDStr := r.URL.Query().Get("action")
	menuIDStr := r.URL.Query().Get("menu_id")

	if mid := strings.TrimSpace(menuIDStr); mid != "" {
		if !menuAccessAllowed(ctx, mid) {
			WebLogf(ctx, "/web", "menu_id=%s denied by access_groups", mid)
			http.Redirect(w, r, "/web/home", http.StatusFound)
			return
		}
	}

	actionID := ResolveWindowActionID(ctx, actionIDStr, menuIDStr)
	if dest, ok := resolveWorkspaceRedirect(ctx, actionIDStr, menuIDStr, actionID); ok {
		if actionID == 0 && dest == "/web/apps" {
			WebLogf(ctx, "/web", "no action for query action=%q menu_id=%q; redirecting to apps", actionIDStr, menuIDStr)
		}
		http.Redirect(w, r, dest, http.StatusFound)
		return
	}

	actionData, err := orm.SearchOne(ctx, "sys.action.window", map[string]interface{}{"id": actionID})
	if err != nil {
		http.Error(w, fmt.Sprintf("Action %d not found", actionID), http.StatusNotFound)
		return
	}

	resolved, err := resolveWorkspaceView(ctx, r, actionData)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "No view for model") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid id") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid id") {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	req := parseWorkspaceRequest(r, actionID)
	viewRecord, err := buildViewRecordData(ctx, w, r, req, resolved, actionData)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid id") {
			status = http.StatusBadRequest
		}
		WebLogf(ctx, "/web", "load view data: %v", err)
		http.Error(w, err.Error(), status)
		return
	}

	html := render.RenderView(ctx, resolved.view, req.menuID, config.AppConfig.TemplatesPath, viewRecord)

	recordID := 0
	if req.recordID != "" {
		if rid, err := strconv.Atoi(req.recordID); err == nil {
			recordID = rid
		}
	}
	WebLogNavigation(ctx, r.URL.Path, "view_open", "Workspace view opened", map[string]interface{}{
		"menu_id":   req.menuID,
		"action_id": actionID,
		"model":     resolved.targetModel,
		"view_type": resolved.selectedMode,
		"record_id": recordID,
		"edit":      req.formEdit,
	})

	writeHTML(w, ctx, r.URL.Path, html)
}
