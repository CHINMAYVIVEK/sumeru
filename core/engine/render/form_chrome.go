package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
)

func renderHeader(ctx context.Context, sb *strings.Builder, h *parser.Header, record map[string]interface{}, vr *ViewRecordData) {
	_ = ctx
	sb.WriteString(`<div class="sum-form-statusbar">`)

	if len(h.Button) > 0 {
		sb.WriteString(`<div class="sum-statusbar-buttons">`)
		model := ""
		id := 0
		next := ""
		if vr != nil {
			model = strings.TrimSpace(vr.ResModel)
			id = vr.RecordID
			if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
				next = "/web?" + q
			}
		}
		for _, b := range h.Button {
			writeObjectActionButton(sb, b, model, id, next, false)
		}
		sb.WriteString(`</div>`)
	}

	sb.WriteString(`<div class="sum-statusbar-status">`)
	writeStatusbarChips(sb, h, record)
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
}

func writeStatusbarChips(sb *strings.Builder, h *parser.Header, record map[string]interface{}) {
	if h == nil {
		return
	}
	for _, hf := range h.Field {
		val := recStr(record, hf.Name)
		if val == "" {
			continue
		}
		sb.WriteString(`<span class="sum-statusbar-chip">`)
		sb.WriteString(template.HTMLEscapeString(val))
		sb.WriteString(`</span>`)
	}
}

func renderButtonBox(sb *strings.Builder, d parser.Div) {
	// Empty button-box: omit spacer chrome (no dead vertical gap).
	if len(d.Field) == 0 && len(d.H1) == 0 {
		return
	}
	sb.WriteString(`<div class="sum-form-toolbar-spacer" aria-hidden="true"></div>`)
}

func renderTitleAvatar(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool, resModel string) {
	_ = ctx
	_ = d
	img := strings.TrimSpace(recStr(record, "image"))
	name := strings.TrimSpace(recStr(record, "name"))
	initials := UserInitialsFromName(name)
	canUpload := fieldDef(resModel, "image") != nil

	hasImg := img != "" && (strings.HasPrefix(img, "http://") || strings.HasPrefix(img, "https://") || strings.HasPrefix(img, "data:"))

	sb.WriteString(`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>`)
	sb.WriteString(`<div class="sum-form-avatar-box sum-form-avatar-box--circle">`)
	if hasImg {
		sb.WriteString(fmt.Sprintf(`<img class="sum-form-avatar-img sum-form-avatar-img--visible" data-sum-avatar-preview src="%s" alt="" />`, template.HTMLEscapeString(img)))
		sb.WriteString(`<span class="sum-form-avatar-initials" data-sum-avatar-initials hidden>` + template.HTMLEscapeString(initials) + `</span>`)
	} else {
		sb.WriteString(`<span class="sum-form-avatar-initials" data-sum-avatar-initials>` + template.HTMLEscapeString(initials) + `</span>`)
		sb.WriteString(`<img class="sum-form-avatar-img" data-sum-avatar-preview alt="" hidden />`)
	}
	sb.WriteString(`</div>`)
	if !ro && canUpload {
		sb.WriteString(fmt.Sprintf(`<input type="hidden" name="image" value="%s" data-sum-avatar-value />`, template.HTMLEscapeString(img)))
		sb.WriteString(`<label class="sum-form-avatar-upload">`)
		sb.WriteString(`<input type="file" accept="image/*" data-sum-avatar-file />`)
		sb.WriteString(`<span>Change</span>`)
		sb.WriteString(`</label>`)
	}
	sb.WriteString(`</div>`)
}

func renderTitleBody(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool) {
	_ = ctx
	sb.WriteString(`<div class="sum-form-title-body sum-form-title-body--main">`)
	roAttr := ""
	if ro {
		roAttr = ` readonly`
	}
	for _, h1 := range d.H1 {
		for _, f := range h1.Field {
			ph := template.HTMLEscapeString(f.Placeholder)
			if ph == "" {
				ph = template.HTMLEscapeString(f.Label)
			}
			if ph == "" {
				ph = "Name"
			}
			v := recStr(record, f.Name)
			sb.WriteString(fmt.Sprintf(`<input class="sum-form-hero-input sum-form-hero-input--bold" placeholder="%s" name="%s" value="%s"%s />`,
				ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v), roAttr))
		}
	}
	sb.WriteString(`<div class="sum-form-contact-row">`)
	for _, f := range d.Field {
		v := recStr(record, f.Name)
		ph := template.HTMLEscapeString(f.Label)
		if ph == "" {
			ph = template.HTMLEscapeString(f.Name)
		}
		inputType := "text"
		if f.Widget == "email" || strings.Contains(strings.ToLower(f.Name), "email") {
			inputType = "email"
		} else if f.Widget == "phone" || f.Widget == "tel" || strings.Contains(strings.ToLower(f.Name), "phone") || strings.Contains(strings.ToLower(f.Name), "mobile") {
			inputType = "tel"
		}
		sb.WriteString(`<div class="sum-form-contact-item">`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-form-inline-input" type="%s" placeholder="%s" name="%s" value="%s"%s />`,
			inputType, ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v), roAttr))
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div></div>`)
}

func renderTitle(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool, resModel string) {
	sb.WriteString(`<div class="sum-form-title-row sum-form-title-row--sheet">`)
	if fieldDef(resModel, "image") != nil {
		renderTitleAvatar(ctx, sb, d, record, ro, resModel)
	}
	renderTitleBody(ctx, sb, d, record, ro)
	sb.WriteString(`</div>`)
}

func renderFormFooter(sb *strings.Builder, ft *parser.Footer, vr *ViewRecordData) {
	if ft == nil || len(ft.Button) == 0 {
		return
	}
	model := ""
	id := 0
	next := ""
	if vr != nil {
		model = strings.TrimSpace(vr.ResModel)
		id = vr.RecordID
		if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
			next = "/web?" + q
		}
	}
	sb.WriteString(`<div class="sum-form-footer" role="group" aria-label="Form actions">`)
	for _, b := range ft.Button {
		writeObjectActionButton(sb, b, model, id, next, true)
	}
	sb.WriteString(`</div>`)
}
