package web

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

type workspaceRequest struct {
	actionID   int
	menuID     string
	viewType   string
	recordID   string
	formEdit   bool
}

func parseWorkspaceRequest(r *http.Request, actionID int) workspaceRequest {
	menuID := r.URL.Query().Get("menu_id")
	recordID := strings.TrimSpace(r.URL.Query().Get("id"))
	return workspaceRequest{
		actionID: actionID,
		menuID:   menuID,
		viewType: strings.TrimSpace(r.URL.Query().Get("view_type")),
		recordID: recordID,
		formEdit: strings.TrimSpace(r.URL.Query().Get("edit")) == "1",
	}
}

// resolveWorkspaceRedirect returns a redirect URL when menu_id has no window action.
func resolveWorkspaceRedirect(ctx context.Context, actionIDStr, menuIDStr string, actionID int) (string, bool) {
	if actionID != 0 {
		return "", false
	}
	if menuIDPointsToAppLogs(ctx, menuIDStr) {
		return "/web/settings/app-logs", true
	}
	if render.IsMenuUnderSettingsRoot(ctx, menuIDStr) {
		return "/web/settings", true
	}
	if isHomeMenuTree(ctx, menuIDStr) {
		dest := "/web/home"
		if mid := strings.TrimSpace(menuIDStr); mid != "" {
			dest = "/web/home?menu_id=" + mid
		}
		return dest, true
	}
	return "/web/apps", true
}

type resolvedWorkspaceView struct {
	viewData     map[string]interface{}
	selectedMode string
	view         *parser.View
	targetModel  string
}

func resolveWorkspaceView(ctx context.Context, r *http.Request, actionData map[string]interface{}) (*resolvedWorkspaceView, error) {
	targetModel := actionWindowTargetModel(actionData)
	if targetModel == "" {
		return nil, fmt.Errorf("action has no target model")
	}

	modes := splitViewModes(strings.TrimSpace(orm.AsString(actionData["view_mode"])))
	if len(modes) == 0 {
		modes = []string{"list"}
	}
	if vt := strings.TrimSpace(r.URL.Query().Get("view_type")); vt != "" {
		modes = append([]string{normalizeViewMode(vt)}, modes...)
	}
	if idStr := strings.TrimSpace(r.URL.Query().Get("id")); idStr != "" {
		if _, err := strconv.Atoi(idStr); err == nil {
			modes = append([]string{"form"}, modes...)
		}
	}

	actionViewID := actionViewIDFromContext(actionData)
	var viewData map[string]interface{}
	var selectedMode string
	var lastErr error
	for _, mode := range modes {
		normalized := normalizeViewMode(mode)
		if normalized == "" {
			continue
		}
		var candidate map[string]interface{}
		var err error
		primaryMode := ""
		if len(modes) > 0 {
			primaryMode = normalizeViewMode(modes[0])
		}
		if actionViewID != "" && normalized == primaryMode {
			candidate, err = orm.FindUIViewByName(ctx, targetModel, normalized, actionViewID)
		}
		if candidate == nil {
			candidate, err = orm.FindUIDefaultView(ctx, targetModel, normalized)
		}
		if err == nil {
			viewData = candidate
			selectedMode = normalized
			break
		}
		lastErr = err
	}
	if viewData == nil {
		msg := fmt.Sprintf("No view for model %s (tried modes: %s)", targetModel, strings.Join(modes, ", "))
		if lastErr != nil && lastErr != sql.ErrNoRows {
			msg = fmt.Sprintf("%s: %v", msg, lastErr)
		}
		return nil, fmt.Errorf("%s", msg)
	}

	arch := strings.TrimSpace(orm.AsString(viewData["arch"]))
	if arch == "" {
		return nil, fmt.Errorf("view arch is empty")
	}
	view, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return nil, fmt.Errorf("parse view arch: %w", err)
	}

	return &resolvedWorkspaceView{
		viewData:     viewData,
		selectedMode: selectedMode,
		view:         view,
		targetModel:  targetModel,
	}, nil
}

func buildViewRecordData(ctx context.Context, w http.ResponseWriter, r *http.Request, req workspaceRequest, resolved *resolvedWorkspaceView, actionData map[string]interface{}) (*render.ViewRecordData, error) {
	vr := &render.ViewRecordData{ActionID: req.actionID}
	vr.CSRFToken = CSRFTokenForRequest(r)
	for _, f := range ConsumePageFlashes(r, w) {
		vr.FlashMessages = append(vr.FlashMessages, render.FlashMessage{
			Kind:  f.Kind,
			Title: f.Title,
			Body:  f.Body,
		})
	}
	vr.ResModel = resolved.targetModel
	vr.FormEditing = req.formEdit
	vr.FormBaseQuery = formBaseQueryValues(req.actionID, req.menuID, "form", req.recordID)
	if req.recordID != "" {
		if rid, err := strconv.Atoi(req.recordID); err == nil && rid > 0 {
			vr.RecordID = rid
		}
	}
	vr.ViewTabs = render.WorkspaceViewTabs(ctx, resolved.targetModel, req.actionID, req.menuID, resolved.selectedMode, req.recordID)

	if err := loadViewModeData(ctx, vr, resolved, actionData, req); err != nil {
		return nil, err
	}
	return vr, nil
}

func loadViewModeData(ctx context.Context, vr *render.ViewRecordData, resolved *resolvedWorkspaceView, actionData map[string]interface{}, req workspaceRequest) error {
	switch resolved.selectedMode {
	case "form":
		if req.recordID == "" {
			return nil
		}
		id, err := strconv.Atoi(req.recordID)
		if err != nil {
			return fmt.Errorf("invalid id")
		}
		rec, err := orm.SearchOne(ctx, resolved.targetModel, map[string]interface{}{"id": id})
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("record %d not found", id)
			}
			return fmt.Errorf("load record: %w", err)
		}
		if resolved.targetModel == "core.user" {
			if cids, err := orm.UserCompanyIDsForUser(ctx, id); err == nil {
				rec["company_ids"] = cids
			}
		}
		vr.Record = rec
	case "list":
		rows, err := orm.SearchLimit(ctx, resolved.targetModel, actionListDomain(ctx, actionData), 500)
		if err != nil {
			return fmt.Errorf("list load: %w", err)
		}
		vr.ListRows = rows
	case "kanban":
		rows, err := orm.SearchLimit(ctx, resolved.targetModel, actionListDomain(ctx, actionData), 200)
		if err != nil {
			return fmt.Errorf("kanban load: %w", err)
		}
		vr.ListRows = rows
		vr.KanbanModel = resolved.targetModel
		if cols, groupField, draggable := render.BuildKanbanColumns(ctx, resolved.view, rows); groupField != "" {
			vr.KanbanColumns = cols
			vr.KanbanGroupField = groupField
			vr.KanbanDraggable = draggable
		}
	}
	return nil
}
