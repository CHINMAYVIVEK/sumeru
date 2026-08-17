package render_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestBuildPivotDataAndRender(t *testing.T) {
	view := &parser.View{
		Model: "crm.lead",
		Type:  "pivot",
		Field: []parser.Field{
			{Name: "stage", PivotType: "row", Label: "Stage"},
			{Name: "team_id", PivotType: "col", Label: "Team"},
			{Name: "expected_revenue", PivotType: "measure", Label: "Revenue"},
		},
	}
	rows := []map[string]interface{}{
		{"stage": "New", "team_id": 1, "expected_revenue": 100.0},
		{"stage": "New", "team_id": 1, "expected_revenue": 50.0},
		{"stage": "Won", "team_id": 2, "expected_revenue": 200.0},
	}
	data := render.BuildPivotData(view, rows)
	if data == nil {
		t.Fatal("nil pivot data")
	}
	if data.Values["New"]["1"] != 150 {
		t.Fatalf("expected 150 for New/1, got %v", data.Values["New"]["1"])
	}
	html := render.RenderPivot(context.Background(), view, data)
	if !strings.Contains(html, "sum-pivot-table") {
		t.Fatalf("expected pivot table html, got %q", html)
	}
}
