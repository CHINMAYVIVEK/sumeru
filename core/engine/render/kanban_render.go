package render

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
	"sumeru/core/report"
)

func kanbanImageOK(src string) bool { return SafeImageSrc(src) }

func isKanbanImageField(f parser.Field) bool {
	return f.Name == "image" || strings.EqualFold(f.Widget, "image") || strings.EqualFold(f.Widget, "circle")
}

func kanbanImageCircle(f parser.Field) bool {
	if strings.EqualFold(f.Widget, "circle") {
		return true
	}
	opt := strings.ToLower(strings.TrimSpace(f.Options))
	if opt == "" {
		return false
	}
	return strings.Contains(opt, "circle") || strings.Contains(opt, `'shape': 'circle'`) || strings.Contains(opt, `"shape": "circle"`)
}

func formatKanbanFieldValue(ctx context.Context, model, name string, row map[string]interface{}) string {
	raw, has := rawField(row, name)
	switch name {
	case "is_company":
		if has && isTruthyDB(raw) {
			return "Company"
		}
		return "Person"
	case "active":
		if !has || isTruthyDB(raw) {
			return "Active"
		}
		return "Inactive"
	case "priority":
		return formatPriorityDisplay(raw)
	default:
		return displayCell(ctx, model, name, row)
	}
}

func formatPriorityDisplay(raw interface{}) string {
	n, ok := orm.CoerceInt64(raw)
	if !ok {
		return ""
	}
	switch n {
	case 3:
		return "Very High"
	case 2:
		return "High"
	case 1:
		return "Medium"
	case 0:
		return "Low"
	default:
		return ""
	}
}

func RenderKanban(ctx context.Context, view *parser.View, recData *ViewRecordData, actionID int, menuID string) string {
	if recData == nil {
		recData = &ViewRecordData{}
	}
	rows := recData.ListRows
	if rows == nil {
		rows = []map[string]interface{}{}
	}

	if len(recData.KanbanColumns) > 0 && recData.KanbanGroupField != "" {
		return renderGroupedKanban(ctx, view, recData.KanbanColumns, recData.KanbanGroupField, recData.KanbanDraggable, recData, actionID, menuID)
	}
	if cols, groupField, draggable := BuildKanbanColumns(ctx, view, rows); groupField != "" && len(cols) > 0 {
		return renderGroupedKanban(ctx, view, cols, groupField, draggable, recData, actionID, menuID)
	}
	return renderFlatKanban(ctx, view, rows, actionID, menuID, recData)
}

func renderGroupedKanban(ctx context.Context, view *parser.View, cols []KanbanColumn, groupField string, draggable bool, recData *ViewRecordData, actionID int, menuID string) string {
	var sb strings.Builder
	renderKanbanReportBar(&sb, view, recData, actionID, menuID)
	model := view.Model
	dragAttr := ""
	if draggable {
		dragAttr = ` data-draggable="1"`
	}
	csrf := template.HTMLEscapeString(recData.CSRFToken)
	sb.WriteString(`<div class="sum-kanban-board sum-kanban-board--grouped"` + dragAttr +
		` data-model="` + template.HTMLEscapeString(model) + `"` +
		` data-group-field="` + template.HTMLEscapeString(groupField) + `"` +
		` data-csrf="` + csrf + `">`)
	sb.WriteString(`<div class="sum-kanban-stage-columns">`)

	for _, col := range cols {
		tooltip := ""
		if col.Tooltip != "" {
			tooltip = ` title="` + template.HTMLEscapeString(col.Tooltip) + `"`
		}
		sb.WriteString(`<section class="sum-kanban-stage-column" data-group-value="` + strconv.FormatInt(col.Value, 10) + `">`)
		sb.WriteString(`<header class="sum-kanban-stage-header"` + tooltip + `>`)
		sb.WriteString(`<span class="sum-kanban-stage-title">` + template.HTMLEscapeString(col.Label) + `</span>`)
		sb.WriteString(`<span class="sum-kanban-stage-count">` + strconv.Itoa(len(col.Records)) + `</span>`)
		sb.WriteString(`</header>`)
		sb.WriteString(`<div class="sum-kanban-cards">`)
		for _, row := range col.Records {
			sb.WriteString(renderKanbanCard(ctx, view, row, actionID, menuID, draggable))
		}
		sb.WriteString(`</div></section>`)
	}
	sb.WriteString(`</div></div>`)
	return sb.String()
}

func renderFlatKanban(ctx context.Context, view *parser.View, rows []map[string]interface{}, actionID int, menuID string, recData *ViewRecordData) string {
	var sb strings.Builder
	renderKanbanReportBar(&sb, view, recData, actionID, menuID)
	sb.WriteString(`<div class="sum-kanban-board">`)
	sb.WriteString(`<div class="sum-kanban-columns">`)
	if len(rows) == 0 {
		sb.WriteString(`<div class="sum-kanban-empty">No records</div>`)
	}
	for _, row := range rows {
		sb.WriteString(renderKanbanCard(ctx, view, row, actionID, menuID, false))
	}
	sb.WriteString(`</div></div>`)
	return sb.String()
}

func renderKanbanCard(ctx context.Context, view *parser.View, row map[string]interface{}, actionID int, menuID string, draggable bool) string {
	rid, ok := orm.CoerceInt64(row["id"])
	if !ok {
		return ""
	}
	href := rowOpenURL(actionID, menuID, rid)

	hasImageField := false
	imageCircle := false
	for _, f := range view.Field {
		if isKanbanImageField(f) {
			hasImageField = true
			if kanbanImageCircle(f) {
				imageCircle = true
			}
			break
		}
	}

	title := ""
	var detailParts []string
	for _, f := range view.Field {
		if isKanbanImageField(f) {
			continue
		}
		val := formatKanbanFieldValue(ctx, view.Model, f.Name, row)
		if f.Name == "name" {
			title = val
			continue
		}
		if val != "" {
			detailParts = append(detailParts, val)
		}
	}
	if title == "" {
		title = recStr(row, "name")
	}
	if title == "" && len(detailParts) > 0 {
		title = detailParts[0]
		detailParts = detailParts[1:]
	}
	if title == "" {
		title = fmt.Sprintf("#%d", rid)
	}

	drag := ""
	rotting := kanbanCardRotting(ctx, row)
	cardClass := "sum-kanban-card"
	if rotting {
		cardClass += " sum-kanban-card--rotting"
	}
	if draggable {
		drag = ` draggable="true"`
	}
	var sb strings.Builder
	sb.WriteString(`<article class="` + cardClass + `"` + drag +
		` data-record-id="` + strconv.FormatInt(rid, 10) + `"` +
		` data-href="` + template.HTMLEscapeString(href) + `"` +
		` role="link" tabindex="0">`)

	if hasImageField {
		img := strings.TrimSpace(recStr(row, "image"))
		initials := UserInitialsFromName(title)
		mediaClass := "sum-kanban-card-media"
		if imageCircle {
			mediaClass += " sum-kanban-card-media--circle"
		} else {
			mediaClass += " sum-kanban-card-media--square"
		}
		sb.WriteString(`<div class="` + mediaClass + `" aria-hidden="true">`)
		if kanbanImageOK(img) {
			sb.WriteString(fmt.Sprintf(`<img class="sum-kanban-card-media-img" src="%s" alt="" />`, template.HTMLEscapeString(img)))
		} else {
			sb.WriteString(`<span class="sum-kanban-card-media-initials">` + template.HTMLEscapeString(initials) + `</span>`)
		}
		sb.WriteString(`</div><div class="sum-kanban-card-body">`)
	}
	sb.WriteString(`<h4 class="sum-kanban-card-title">` + template.HTMLEscapeString(title) + `</h4>`)
	if stars := kanbanPriorityStars(view, row); stars != "" {
		sb.WriteString(stars)
	}
	if len(detailParts) > 0 {
		sb.WriteString(`<p class="sum-kanban-card-sub">` + template.HTMLEscapeString(strings.Join(detailParts, " · ")) + `</p>`)
	}
	if hasImageField {
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</article>`)
	return sb.String()
}

func kanbanPriorityStars(view *parser.View, row map[string]interface{}) string {
	for _, f := range view.Field {
		if f.Widget != "priority" && f.Name != "priority" {
			continue
		}
		n, ok := orm.CoerceInt64(row[f.Name])
		if !ok {
			n = 1
		}
		var sb strings.Builder
		sb.WriteString(`<div class="sum-kanban-priority" aria-label="Priority">`)
		for i := int64(0); i <= 3; i++ {
			cls := "sum-kanban-priority-star"
			if i <= n {
				cls += " sum-kanban-priority-star--on"
			}
			sb.WriteString(`<span class="` + cls + `">★</span>`)
		}
		sb.WriteString(`</div>`)
		return sb.String()
	}
	return ""
}

func kanbanCardRotting(ctx context.Context, row map[string]interface{}) bool {
	stageID, ok := orm.CoerceInt64(row["stage_id"])
	if !ok || stageID <= 0 {
		return false
	}
	stage, err := orm.SearchOne(ctx, "crm.stage", map[string]interface{}{"id": stageID})
	if err != nil {
		return false
	}
	threshold, ok := orm.CoerceInt64(stage["rotting_threshold_days"])
	if !ok || threshold <= 0 {
		return false
	}
	last := strings.TrimSpace(orm.AsString(row["date_last_stage_update"]))
	if last == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, last)
	if err != nil {
		if t2, err2 := time.Parse("2006-01-02", last); err2 == nil {
			t = t2
		} else {
			return false
		}
	}
	return time.Since(t) > time.Duration(threshold)*24*time.Hour
}

func renderKanbanReportBar(sb *strings.Builder, view *parser.View, recData *ViewRecordData, actionID int, menuID string) {
	caps := report.CapabilitiesFromView(view)
	if !caps.HasDownload() && !caps.BulkUpload {
		return
	}
	csrf := ""
	if recData != nil {
		csrf = recData.CSRFToken
	}
	sb.WriteString(`<div class="sum-list-control sum-kanban-report-bar">`)
	sb.WriteString(`<div class="sum-list-control-left">`)
	sb.WriteString(RenderReportToolbar(caps, view.Model, actionID, menuID, 0, view.Field, csrf))
	sb.WriteString(`</div></div>`)
}
