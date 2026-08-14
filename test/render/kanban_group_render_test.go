package render_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestRenderGroupedKanbanColumns(t *testing.T) {
	v := &parser.View{
		Model:          "crm.lead",
		Type:           "kanban",
		DefaultGroupBy: "stage_id",
		Field:          []parser.Field{{Name: "name"}, {Name: "expected_revenue"}},
	}
	rec := &render.ViewRecordData{
		CSRFToken: "tok",
		ListRows: []map[string]interface{}{
			{"id": int64(1), "name": "Lead A", "stage_id": int64(10)},
			{"id": int64(2), "name": "Lead B", "stage_id": int64(20)},
		},
		KanbanColumns: []render.KanbanColumn{
			{Value: 10, Label: "New", Records: []map[string]interface{}{{"id": int64(1), "name": "Lead A", "stage_id": int64(10)}}},
			{Value: 20, Label: "Won", Records: []map[string]interface{}{{"id": int64(2), "name": "Lead B", "stage_id": int64(20)}}},
		},
		KanbanGroupField: "stage_id",
		KanbanDraggable:  true,
	}
	html := render.RenderKanban(context.Background(), v, rec, 5, "crm_menu")
	if !strings.Contains(html, "sum-kanban-board--grouped") {
		t.Fatal("expected grouped board")
	}
	if !strings.Contains(html, `data-group-value="10"`) {
		t.Fatal("expected stage column")
	}
	if !strings.Contains(html, `draggable="true"`) {
		t.Fatal("expected draggable cards")
	}
	if !strings.Contains(html, "Lead A") {
		t.Fatal("expected card content")
	}
}
