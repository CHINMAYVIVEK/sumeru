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

func kanbanImageOK(src string) bool {
	return src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:"))
}

func isKanbanImageField(f parser.Field) bool {
	return f.Name == "image" || strings.EqualFold(f.Widget, "image") || strings.EqualFold(f.Widget, "circle")
}

// kanbanImageCircle is true when XML explicitly requests a circular photo
// via widget="circle" or options containing shape=circle / "circle".
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
	default:
		return displayCell(ctx, model, name, row)
	}
}

// RenderKanban renders a simple kanban board for workspace views.
func RenderKanban(ctx context.Context, view *parser.View, rows []map[string]interface{}, actionID int, menuID string) string {
	if rows == nil {
		rows = []map[string]interface{}{}
	}
	var sb strings.Builder

	sb.WriteString(`<div class="sum-kanban-board">`)
	sb.WriteString(`<div class="sum-kanban-columns">`)
	if len(rows) == 0 {
		sb.WriteString(`<div class="sum-kanban-empty">No records</div>`)
	}

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

	for _, row := range rows {
		rid, ok := orm.CoerceInt64(row["id"])
		if !ok {
			continue
		}
		href := rowOpenURL(actionID, menuID, rid)
		qhref := strconv.Quote(href)

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

		sb.WriteString(`<article class="sum-kanban-card" role="link" tabindex="0" onclick='window.location.href=` + qhref + `' onkeydown='if(event.key==="Enter"){window.location.href=` + qhref + `}'>`)
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
			sb.WriteString(`</div>`)
			sb.WriteString(`<div class="sum-kanban-card-body">`)
		}
		sb.WriteString(`<h4 class="sum-kanban-card-title">` + template.HTMLEscapeString(title) + `</h4>`)
		if len(detailParts) > 0 {
			sb.WriteString(`<p class="sum-kanban-card-sub">` + template.HTMLEscapeString(strings.Join(detailParts, " · ")) + `</p>`)
		}
		if hasImageField {
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</article>`)
	}
	sb.WriteString(`</div></div>`)

	return sb.String()
}
