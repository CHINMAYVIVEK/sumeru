package swcmeta

import (
	"context"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// ViewRecordInput is workspace ORM data passed from the web layer (no render import).
type ViewRecordInput struct {
	ActionID        int
	ResModel        string
	RecordID        int
	FormEditing     bool
	CSRFToken       string
	FormBaseQuery   string
	ListSearchQuery string
	ListSearchURL   string
	Record          map[string]interface{}
	ListRows        []map[string]interface{}
	KanbanColumns   []KanbanColumn
	KanbanGroupField string
	KanbanDraggable bool
	Pivot           *PivotMeta
	ViewTabs        []ViewTab
	Breadcrumbs     []Breadcrumb
}

// BuildWorkspacePayload serializes loaded workspace data for SWC.
func BuildWorkspacePayload(
	ctx context.Context,
	view *parser.View,
	selectedMode string,
	rec ViewRecordInput,
	reqMenuID string,
) WorkspacePayload {
	arch := SerializeView(view)
	arch.Type = selectedMode

	if len(rec.KanbanColumns) > 0 || rec.KanbanGroupField != "" {
		arch.Kanban = &KanbanMeta{
			GroupField: rec.KanbanGroupField,
			Draggable:  rec.KanbanDraggable,
			Columns:    rec.KanbanColumns,
		}
	}
	if rec.Pivot != nil {
		arch.Pivot = rec.Pivot
	}

	payload := WorkspacePayload{
		ActionID:      rec.ActionID,
		MenuID:        reqMenuID,
		ViewType:      selectedMode,
		Model:         rec.ResModel,
		RecordID:      rec.RecordID,
		FormEdit:      rec.FormEditing,
		CSRFToken:     rec.CSRFToken,
		Arch:          arch,
		ListSearch:    rec.ListSearchQuery,
		ListSearchURL: rec.ListSearchURL,
		FormBaseQuery: rec.FormBaseQuery,
		ViewTabs:      rec.ViewTabs,
		Breadcrumbs:   rec.Breadcrumbs,
	}

	if rec.Record != nil {
		payload.Record = redactCopy(ctx, rec.ResModel, rec.Record)
	}
	if len(rec.ListRows) > 0 {
		payload.Records = redactRows(ctx, rec.ResModel, rec.ListRows)
	}
	return payload
}

func redactCopy(ctx context.Context, model string, rec map[string]interface{}) map[string]interface{} {
	copy := map[string]interface{}{}
	for k, v := range rec {
		copy[k] = v
	}
	uid := orm.UIDFromContext(ctx)
	orm.RedactRecordForRead(ctx, uid, model, copy)
	return copy
}

func redactRows(ctx context.Context, model string, rows []map[string]interface{}) []map[string]interface{} {
	uid := orm.UIDFromContext(ctx)
	out := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		copy := map[string]interface{}{}
		for k, v := range row {
			copy[k] = v
		}
		orm.RedactRecordForRead(ctx, uid, model, copy)
		out[i] = copy
	}
	return out
}
