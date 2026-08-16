package render

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sumeru/core/report"
)

func RenderList(ctx context.Context, view *parser.View, rows []map[string]interface{}, actionID int, menuID string, csrfToken string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder
	n := len(rows)
	listTitle := HumanViewBreadcrumb(view.Model, "list")
	newHref := formNewRecordURL(actionID, menuID)

	sb.WriteString(`<div class="sum-list-view">`)
	sb.WriteString(`<div class="sum-list-control">`)
	sb.WriteString(`<div class="sum-list-control-left">`)
	sb.WriteString(`<a href="` + template.HTMLEscapeString(newHref) + `" class="sum-list-btn-new">New</a>`)
	caps := report.CapabilitiesFromView(view)
	sb.WriteString(RenderReportToolbar(caps, view.Model, actionID, menuID, 0, view.Field, csrfToken))
	sb.WriteString(`<h1 class="sum-list-title">` + template.HTMLEscapeString(listTitle) + `</h1>`)
	sb.WriteString(`<button type="button" class="sum-list-icon-btn" disabled aria-hidden="true" title="Configuration">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>`)
	sb.WriteString(`</button>`)
	if n > 0 {
		sb.WriteString(fmt.Sprintf(`<span class="sum-list-pager" aria-live="polite">1-%d / %d</span>`, n, n))
	} else {
		sb.WriteString(`<span class="sum-list-pager">0 / 0</span>`)
	}
	sb.WriteString(`</div>`)
	sb.WriteString(`<div class="sum-list-search-wrap" role="search">`)
	sb.WriteString(`<span class="sum-list-search-icon" aria-hidden="true"><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="M21 21l-4.35-4.35"/></svg></span>`)
	sb.WriteString(`<input type="search" class="sum-list-search" placeholder="Search…" disabled aria-disabled="true" />`)
	sb.WriteString(`<span class="sum-list-search-caret" aria-hidden="true"></span>`)
	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="sum-web-table-wrap sum-list-table-wrap">`)
	sb.WriteString(`<table class="sum-list-table">`)
	sb.WriteString(`<thead><tr>`)
	sb.WriteString(`<th class="sum-list-th-check" scope="col"><span class="sum-list-th-muted"><input type="checkbox" disabled aria-label="Select all" /></span></th>`)
	sb.WriteString(`<th class="sum-list-th-grip" scope="col" aria-hidden="true"></th>`)
	for _, f := range view.Field {
		label := FieldDisplayLabel(f)
		sb.WriteString(`<th class="sum-list-th">` + template.HTMLEscapeString(label) + `</th>`)
	}
	sb.WriteString(`<th class="sum-list-th-actions" scope="col" aria-label="Columns"><button type="button" class="sum-list-icon-btn sum-list-icon-btn--ghost" disabled aria-hidden="true">`)
	sb.WriteString(`<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 6h16M4 12h16M4 18h16"/></svg>`)
	sb.WriteString(`</button></th>`)
	sb.WriteString(`</tr></thead>`)
	sb.WriteString(`<tbody>`)
	colspan := len(view.Field) + 3
	if colspan < 3 {
		colspan = 3
	}
	if len(rows) == 0 {
		sb.WriteString(fmt.Sprintf(`<tr><td colspan="%d" class="sum-list-empty">No records</td></tr>`, colspan))
	}
	for _, row := range rows {
		rid, hasID := orm.CoerceInt64(row["id"])
		menuTrim := strings.TrimSpace(menuID)
		canOpenForm := !view.ListNoRowOpen && hasID && rid > 0 && (actionID > 0 || menuTrim != "")
		rowClass := "sum-list-row"
		if canOpenForm {
			rowClass += " sum-list-row--click"
		}
		sb.WriteString(`<tr class="` + rowClass + `"`)
		if canOpenForm {
			href := rowOpenURL(actionID, menuID, rid)
			qhref := strconv.Quote(href)
			sb.WriteString(` role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"||event.key===" "){event.preventDefault();window.location.href=` + qhref + `}'`)
		}
		sb.WriteString(`>`)
		sb.WriteString(`<td class="sum-list-td-check"><input type="checkbox" disabled onclick="event.stopPropagation()" aria-label="Select row" /></td>`)
		sb.WriteString(`<td class="sum-list-td-grip" aria-hidden="true"><span class="sum-list-grip">⠿</span></td>`)
		for _, f := range view.Field {
			cell := template.HTMLEscapeString(displayCell(ctx, view.Model, f.Name, row))
			sb.WriteString(`<td class="sum-list-td">` + cell + `</td>`)
		}
		sb.WriteString(`<td class="sum-list-td-actions"></td>`)
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table></div></div>`)

	return sb.String()
}
