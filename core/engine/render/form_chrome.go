package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
)

func renderHeader(ctx context.Context, sb *strings.Builder, h *parser.Header, record map[string]interface{}) {
	_ = ctx
	sb.WriteString(`<div class="sum-form-statusbar">`)

	sb.WriteString(`<div class="sum-statusbar-buttons">`)
	for _, b := range h.Button {
		class := "sum-header-btn "
		if b.Class == "sum_highlight" {
			class += "sum-header-btn--primary"
		} else {
			class += "sum-header-btn--secondary"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" disabled class="%s sum-header-btn--disabled">%s</button>`, class, template.HTMLEscapeString(b.String)))
	}
	sb.WriteString(`</div>`)

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
	_ = d
	sb.WriteString(`<div class="sum-form-toolbar-spacer" aria-hidden="true"></div>`)
}

func renderTitle(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool) {
	_ = ctx
	sb.WriteString(`<div class="sum-form-title-row sum-form-title-row--sheet">`)

	sb.WriteString(`<div class="sum-form-avatar">`)
	sb.WriteString(`<div class="sum-form-avatar-box">`)
	sb.WriteString(`<svg class="sum-form-avatar-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"></path><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z"></path></svg>`)
	sb.WriteString(`<span class="sum-form-avatar-hint">Your logo</span>`)
	sb.WriteString(`</div></div>`)

	sb.WriteString(`<div class="sum-form-title-body">`)
	if ro {
		for _, h1 := range d.H1 {
			for _, f := range h1.Field {
				ph := strings.TrimSpace(f.Label)
				if ph == "" {
					ph = "Title"
				}
				v := recStr(record, f.Name)
				sb.WriteString(`<div class="sum-read-hero-kicker">` + template.HTMLEscapeString(ph) + `</div>`)
				sb.WriteString(`<div class="sum-read-hero-title">` + template.HTMLEscapeString(v) + `</div>`)
			}
		}
		sb.WriteString(`<div class="sum-read-meta-stack">`)
		for _, f := range d.Field {
			icon := "M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
			if strings.Contains(f.Name, "phone") {
				icon = "M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
			}
			if strings.Contains(f.Name, "tag") {
				icon = "M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
			}
			v := recStr(record, f.Name)
			sb.WriteString(`<div class="sum-read-inline">`)
			sb.WriteString(fmt.Sprintf(`<svg class="sum-read-inline-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="%s"></path></svg>`, icon))
			sb.WriteString(`<span class="sum-read-inline-text">` + template.HTMLEscapeString(v) + `</span>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	} else {
		for _, h1 := range d.H1 {
			for _, f := range h1.Field {
				ph := template.HTMLEscapeString(f.Label)
				v := recStr(record, f.Name)
				sb.WriteString(fmt.Sprintf(`<input class="sum-form-hero-input" placeholder="%s" name="%s" value="%s" />`,
					ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v)))
			}
		}
		sb.WriteString(`<div class="sum-form-inline-fields">`)
		for _, f := range d.Field {
			icon := "M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
			if strings.Contains(f.Name, "phone") {
				icon = "M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z"
			}
			if strings.Contains(f.Name, "tag") {
				icon = "M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z"
			}
			v := recStr(record, f.Name)
			sb.WriteString(`<div class="sum-form-inline-row">`)
			sb.WriteString(fmt.Sprintf(`<svg class="sum-form-inline-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="%s"></path></svg>`, icon))
			sb.WriteString(fmt.Sprintf(`<input class="sum-form-inline-input" placeholder="%s" name="%s" value="%s" />`,
				template.HTMLEscapeString(f.Label), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v)))
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}

	sb.WriteString(`</div>`)
	sb.WriteString(`</div>`)
}

func renderFormFooter(sb *strings.Builder, ft *parser.Footer) {
	if ft == nil || len(ft.Button) == 0 {
		return
	}
	sb.WriteString(`<div class="sum-form-footer" role="group" aria-label="Form actions">`)
	for _, b := range ft.Button {
		btnClass := "sum-form-footer-btn"
		if strings.Contains(b.Class, "sum_highlight") || strings.Contains(b.Class, "btn-primary") {
			btnClass += " sum-form-footer-btn--primary"
		}
		label := template.HTMLEscapeString(b.String)
		if label == "" {
			label = template.HTMLEscapeString(b.Name)
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" disabled name="%s" class="%s" aria-disabled="true">%s</button>`,
			template.HTMLEscapeString(b.Name), btnClass, label))
	}
	sb.WriteString(`</div>`)
}
