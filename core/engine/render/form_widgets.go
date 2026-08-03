package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func renderField(ctx context.Context, sb *strings.Builder, f parser.Field, record map[string]interface{}, ro bool) {
	if gs := strings.TrimSpace(f.Groups); gs != "" {
		if !orm.UserHasAnyAccessGroup(ctx, orm.SecurityUID(ctx), gs) {
			return
		}
	}
	label := f.Label
	if label == "" {
		label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
	}

	raw, hasRaw := rawField(record, f.Name)
	isBoolish := f.Widget == "boolean" || strings.HasSuffix(f.Name, "_active")
	if isBoolish {
		checked := ""
		if hasRaw && isTruthyDB(raw) {
			checked = ` checked`
		}
		dis := ""
		if ro {
			dis = ` disabled`
		}
		if ro {
			val := "No"
			if hasRaw && isTruthyDB(raw) {
				val = "Yes"
			}
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value sum-read-value--bool">` + template.HTMLEscapeString(val) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-field-checkbox" id="%s" name="%s" type="checkbox"%s%s />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), checked, dis))
		sb.WriteString(`</div>`)
		return
	}

	if f.Widget == "image" {
		src := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">`)
			if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
				sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-read-image" />`, template.HTMLEscapeString(src)))
			} else {
				sb.WriteString(`<span class="sum-read-empty">—</span>`)
			}
			sb.WriteString(`</div></div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
			sb.WriteString(`<div class="sum-image-thumb">`)
			sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-image-thumb-img" />`, template.HTMLEscapeString(src)))
			sb.WriteString(`</div>`)
		} else {
			sb.WriteString(`<div class="sum-image-placeholder">`)
			sb.WriteString(`<svg class="sum-image-placeholder-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	} else if f.Widget == "many2many_tags" {
		txt := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(txt) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(`<div class="sum-field-tags">`)
		sb.WriteString(template.HTMLEscapeString(txt))
		sb.WriteString(`</div></div>`)
	} else {
		placeholder := f.Placeholder
		if placeholder == "" {
			placeholder = "Enter " + strings.ToLower(label) + "..."
		}
		val := recStr(record, f.Name)
		if ro {
			display := val
			if display == "" {
				display = "—"
			}
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(display) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" name="%s" type="text" placeholder="%s" value="%s" />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
			template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val)))
		sb.WriteString(`</div>`)
	}
}
