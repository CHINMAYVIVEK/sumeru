package render_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestRenderList_rowClickUsesValidOnclickQuotes(t *testing.T) {
	v := &parser.View{Type: "list", Field: []parser.Field{{Name: "name"}}}
	rows := []map[string]interface{}{{"id": 1, "name": "Acme"}}
	html := render.RenderList(context.Background(), v, rows, 12, "3", "tok", "", "/web?action=12&menu_id=3&view_type=list")
	if !strings.Contains(html, "onclick='window.location.href=") {
		t.Fatalf("expected row onclick with single-quoted attribute, got: %s", html)
	}
}
