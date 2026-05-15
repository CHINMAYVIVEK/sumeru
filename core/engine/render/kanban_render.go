package render

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// RenderKanban renders a simple kanban board for workspace views.
func RenderKanban(_ context.Context, view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder

	sb.WriteString(`<div class="sum-kanban-board">`)
	sb.WriteString(`<div class="sum-kanban-columns">`)
	if len(rows) == 0 {
		sb.WriteString(`<div class="sum-kanban-empty">No records</div>`)
	}
	for _, row := range rows {
		rid, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		href := rowOpenURL(actionID, menuID, rid)
		title := ""
		if len(view.Field) > 0 {
			title = recStr(row, view.Field[0].Name)
		}
		if title == "" {
			title = recStr(row, "name")
		}
		if title == "" {
			title = fmt.Sprintf("#%d", rid)
		}
		qhref := strconv.Quote(href)
		sb.WriteString(`<article class="sum-kanban-card" role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"){window.location.href=` + qhref + `}'>`)
		sb.WriteString(`<h4 class="sum-kanban-card-title">` + template.HTMLEscapeString(title) + `</h4>`)
		var subParts []string
		for fi := 1; fi < len(view.Field); fi++ {
			s := recStr(row, view.Field[fi].Name)
			if s != "" {
				subParts = append(subParts, s)
			}
		}
		if len(subParts) > 0 {
			sb.WriteString(`<p class="sum-kanban-card-sub">` + template.HTMLEscapeString(strings.Join(subParts, " · ")) + `</p>`)
		}
		sb.WriteString(`</article>`)
	}
	sb.WriteString(`</div></div>`)

	return sb.String()
}
