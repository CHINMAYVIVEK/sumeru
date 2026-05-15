package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// WebHandler renders sys.action.window targets using sys.view definitions.
func WebHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	actionIDStr := r.URL.Query().Get("action")
	menuIDStr := r.URL.Query().Get("menu_id")

	actionID := ResolveWindowActionID(r.Context(), actionIDStr, menuIDStr)

	if actionID == 0 {
		if menuIDPointsToAppLogs(r.Context(), menuIDStr) {
			http.Redirect(w, r, "/web/settings/app-logs", http.StatusFound)
			return
		}
		// Avoid picking an arbitrary window action (often the lowest id); that misroutes e.g. Sales/CRM roots.
		WebLogf("/web", "no action for query action=%q menu_id=%q; redirecting to home", actionIDStr, menuIDStr)
		http.Redirect(w, r, "/web/home", http.StatusFound)
		return
	}

	actionData, err := orm.SearchOne(r.Context(), "sys.action.window", map[string]interface{}{"id": actionID})
	if err != nil {
		http.Error(w, fmt.Sprintf("Action %d not found", actionID), http.StatusNotFound)
		return
	}

	targetModel := actionWindowTargetModel(actionData)
	viewMode := strings.TrimSpace(orm.AsString(actionData["view_mode"]))
	if targetModel == "" {
		http.Error(w, "Action has no target model", http.StatusInternalServerError)
		return
	}

	modes := splitViewModes(viewMode)
	if len(modes) == 0 {
		modes = []string{"tree"}
	}

	// Optional ?view_type=... (explicit tab / layout) is applied before the default mode list.
	if vt := strings.TrimSpace(r.URL.Query().Get("view_type")); vt != "" {
		modes = append([]string{normalizeViewMode(vt)}, modes...)
	}

	// Opening a specific record should use the form view; prepend after view_type so ?id=… always wins over ?view_type=tree|kanban.
	if idStr := strings.TrimSpace(r.URL.Query().Get("id")); idStr != "" {
		if _, err := strconv.Atoi(idStr); err == nil {
			modes = append([]string{"form"}, modes...)
		}
	}

	var viewData map[string]interface{}
	var selectedMode string
	var lastErr error
	for _, mode := range modes {
		nm := normalizeViewMode(mode)
		if nm == "" {
			continue
		}
		vd, err := orm.FindUIDefaultView(r.Context(), targetModel, nm)
		if err == nil {
			viewData = vd
			selectedMode = nm
			break
		}
		lastErr = err
	}
	if viewData == nil {
		msg := fmt.Sprintf("No view for model %s (tried modes: %s)", targetModel, strings.Join(modes, ", "))
		if lastErr != nil && lastErr != sql.ErrNoRows {
			msg = fmt.Sprintf("%s: %v", msg, lastErr)
		}
		http.Error(w, msg, http.StatusNotFound)
		return
	}

	arch := orm.AsString(viewData["arch"])
	if strings.TrimSpace(arch) == "" {
		http.Error(w, "View arch is empty", http.StatusInternalServerError)
		return
	}

	view, err := parser.ParseViewFromArch(arch)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error parsing view architecture: %v", err), http.StatusInternalServerError)
		return
	}

	menuID := r.URL.Query().Get("menu_id")
	idQ := strings.TrimSpace(r.URL.Query().Get("id"))
	editing := strings.TrimSpace(r.URL.Query().Get("edit")) == "1"

	vr := &render.ViewRecordData{ActionID: actionID}
	// ResModel + RecordID drive readonly vs edit (?edit=1) for workspace forms; keep both in sync whenever loading a record.
	vr.ResModel = targetModel
	vr.FormEditing = editing
	vr.FormBaseQuery = formBaseQueryValues(actionID, menuID, "form", idQ)
	if idQ != "" {
		if rid, err := strconv.Atoi(idQ); err == nil && rid > 0 {
			vr.RecordID = rid
		}
	}
	vr.ViewTabs = render.WorkspaceViewTabs(r.Context(), targetModel, actionID, menuID, selectedMode, idQ)

	switch selectedMode {
	case "form":
		idStr := strings.TrimSpace(r.URL.Query().Get("id"))
		if idStr != "" {
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid id", http.StatusBadRequest)
				return
			}
			rec, err := orm.SearchOne(r.Context(), targetModel, map[string]interface{}{"id": id})
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, fmt.Sprintf("Record %d not found", id), http.StatusNotFound)
					return
				}
				WebLogf("/web", "load record %s id=%d: %v", targetModel, id, err)
				http.Error(w, "Failed to load record", http.StatusInternalServerError)
				return
			}
			vr.Record = rec
		}
	case "tree", "list":
		rows, err := orm.SearchLimit(r.Context(), targetModel, nil, 500)
		if err != nil {
			WebLogf("/web", "list %s: %v", targetModel, err)
			http.Error(w, "Failed to load records", http.StatusInternalServerError)
			return
		}
		vr.ListRows = rows
	case "kanban":
		rows, err := orm.SearchLimit(r.Context(), targetModel, nil, 200)
		if err != nil {
			WebLogf("/web", "kanban %s: %v", targetModel, err)
			http.Error(w, "Failed to load records", http.StatusInternalServerError)
			return
		}
		vr.ListRows = rows
	}

	html := render.RenderView(r.Context(), view, menuID, config.AppConfig.TemplatesPath, vr)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
