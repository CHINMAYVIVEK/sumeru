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

// RenderTree renders list/tree HTML for workspace views.
func RenderTree(ctx context.Context, view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder
	n := len(rows)
	listTitle := HumanViewBreadcrumb(view.Model, "tree")
	newHref := formNewRecordURL(actionID, menuID)

	sb.WriteString(`<div class="sum-tree-view">`)
	sb.WriteString(`<div class="sum-tree-control">`)
	sb.WriteString(`<div class="sum-tree-control-left">`)
	sb.WriteString(`<a href="` + template.HTMLEscapeString(newHref) + `" class="sum-tree-btn-new">New</a>`)
	sb.WriteString(`<h1 class="sum-tree-title">` + template.HTMLEscapeString(listTitle) + `</h1>`)
	sb.WriteString(`<button type="button" class="sum-tree-icon-btn" disabled aria-hidden="true" title="Configuration">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`)
	sb.WriteString(`</button>`)
	if n > 0 {
		sb.WriteString(fmt.Sprintf(`<span class="sum-tree-pager" aria-live="polite">1-%d / %d</span>`, n, n))
	} else {
		sb.WriteString(`<span class="sum-tree-pager">0 / 0</span>`)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="sum-tree-search-wrap" role="search">`)
	sb.WriteString(`<span class="sum-tree-search-icon" aria-hidden="true"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="M21 21l-4.35-4.35"/></svg></span>`)
	sb.WriteString(`<input type="search" class="sum-tree-search" placeholder="Search…" disabled aria-disabled="true" />`)
	sb.WriteString(`<span class="sum-tree-search-caret" aria-hidden="true"></span>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="sum-web-table-wrap sum-tree-table-wrap">`)
	sb.WriteString(`<table class="sum-tree-table">`)
	sb.WriteString(`<thead><tr>`)
	sb.WriteString(`<th class="sum-tree-th-check" scope="col"><span class="sum-tree-th-muted"><input type="checkbox" disabled aria-label="Select all" /></span></th>`)
	sb.WriteString(`<th class="sum-tree-th-grip" scope="col" aria-hidden="true"></th>`)
	for _, f := range view.Field {
		label := f.Label
		if label == "" {
			label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
		}
		sb.WriteString(`<th class="sum-tree-th">` + template.HTMLEscapeString(label) + `</th>`)
	}
	sb.WriteString(`<th class="sum-tree-th-actions" scope="col" aria-label="Columns"><button type="button" class="sum-tree-icon-btn sum-tree-icon-btn--ghost" disabled aria-hidden="true">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6h16M4 12h16M4 18h16"/></svg>`)
	sb.WriteString(`</button></th>`)
	sb.WriteString(`</tr></thead>`)
	sb.WriteString(`<tbody>`)
	colspan := len(view.Field) + 3
	if colspan < 3 {
		colspan = 3
	}
	if len(rows) == 0 {
		sb.WriteString(fmt.Sprintf(`<tr><td colspan="%d" class="sum-tree-empty">No records</td></tr>`, colspan))
	}
	for _, row := range rows {
		rid, hasID := orm.CoerceInt64(row["id"])
		menuTrim := strings.TrimSpace(menuID)
		canOpenForm := !view.TreeNoRowOpen && hasID && rid > 0 && (actionID > 0 || menuTrim != "")
		rowClass := "sum-tree-row"
		if canOpenForm {
			rowClass += " sum-tree-row--click"
		}
		sb.WriteString(`<tr class="` + rowClass + `"`)
		if canOpenForm {
			href := rowOpenURL(actionID, menuID, rid)
			qhref := strconv.Quote(href)
			sb.WriteString(` role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"||event.key===" "){event.preventDefault();window.location.href=` + qhref + `}'`)
		}
		sb.WriteString(`>`)
		sb.WriteString(`<td class="sum-tree-td-check"><input type="checkbox" disabled onclick="event.stopPropagation()" aria-label="Select row" /></td>`)
		sb.WriteString(`<td class="sum-tree-td-grip" aria-hidden="true"><span class="sum-tree-grip">⠿</span></td>`)
		for _, f := range view.Field {
			cell := template.HTMLEscapeString(displayCell(ctx, view.Model, f.Name, row))
			sb.WriteString(`<td class="sum-tree-td">` + cell + `</td>`)
		}
		sb.WriteString(`<td class="sum-tree-td-actions"></td>`)
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table></div></div>`)

	return sb.String()
}
