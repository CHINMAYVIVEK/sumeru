package render_test

import (
	"context"
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
)

func TestRenderTree_rowClickUsesValidOnclickQuotes(t *testing.T) {
	v := &parser.View{Type: "tree", Field: []parser.Field{{Name: "name"}}}
	rows := []map[string]interface{}{{"id": int64(7), "name": "Acme"}}
	html := render.RenderTree(context.Background(), v, rows, 12, "3")
	if strings.Contains(html, `onclick="window.location.href="/`) {
		t.Fatal("onclick uses invalid nested double quotes")
	}
	if !strings.Contains(html, `onclick='window.location.href="/web?`) {
		t.Fatalf("expected onclick with outer single quotes and JS double-quoted URL; got: %s", html)
	}
}
