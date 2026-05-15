package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// RenderForm renders a parsed form view into HTML (sheet, header, fields, notebooks).
func RenderForm(ctx context.Context, view *parser.View, vr *ViewRecordData) string {
	if vr == nil {
		vr = &ViewRecordData{}
	}
	record := vr.Record
	if record == nil {
		record = map[string]interface{}{}
	}
	ro := formFieldReadonly(vr)
	chrome := workspaceFormChrome(vr)

	var sb strings.Builder

	formViewClass := "sum-form-view"
	if ro && vr.RecordID > 0 {
		formViewClass += " sum-form-view--readonly"
	}
	if chrome {
		formViewClass += " sum-form-view--workspace-chrome"
	}
	sb.WriteString(`<div class="` + formViewClass + `">`)

	sb.WriteString(`<div class="sum-form-sheet-bg">`)

	if chrome {
		renderWorkspaceFormToolbar(&sb, vr, view.Header, record)
	} else if view.Header != nil {
		renderHeader(ctx, &sb, view.Header, record)
	}

	if view.Sheet != nil {
		renderSheet(ctx, &sb, view.Sheet, record, ro, vr)
	} else {
		sb.WriteString(`<div class="sum-form-sheet sum-form-sheet--solo">`)
		for _, f := range view.Field {
			renderField(ctx, &sb, f, record, ro)
		}
		for _, g := range view.Group {
			renderGroup(ctx, &sb, g, record, ro)
		}
		sb.WriteString(`</div>`)
	}

	if view.Footer != nil {
		renderFormFooter(&sb, view.Footer)
	}

	if chrome {
		renderWorkspaceFormChromeClose(&sb, vr)
	}

	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)

	return sb.String()
}

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

func renderSheet(ctx context.Context, sb *strings.Builder, s *parser.Sheet, record map[string]interface{}, ro bool, vr *ViewRecordData) {
	sb.WriteString(`<div class="sum-form-sheet">`)

	for _, d := range s.Div {
		if strings.Contains(d.Class, "sum_button_box") {
			renderButtonBox(sb, d)
		}
	}

	hasSumTitle := false
	for _, d := range s.Div {
		if strings.Contains(d.Class, "sum_title") {
			hasSumTitle = true
			break
		}
	}

	if hasSumTitle {
		sb.WriteString(`<div class="sum-form-split-layout" data-sum-form-split>`)
		sb.WriteString(`<aside class="sum-form-split-left" aria-label="Record summary">`)
		for _, d := range s.Div {
			if strings.Contains(d.Class, "sum_title") {
				renderTitle(ctx, sb, d, record, ro)
			}
		}
		sb.WriteString(`</aside>`)
		sb.WriteString(`<div class="sum-form-split-resizer" role="separator" aria-orientation="vertical" aria-label="Resize columns" tabindex="0"></div>`)
		sb.WriteString(`<div class="sum-form-split-main">`)
	} else {
		for _, d := range s.Div {
			if strings.Contains(d.Class, "sum_title") {
				renderTitle(ctx, sb, d, record, ro)
			}
		}
	}

	if ro {
		sb.WriteString(`<div class="sum-read-fields sum-read-fields--sheet sum-field-region--sheet">`)
	} else {
		sb.WriteString(`<div class="sum-form-edit-grid sum-field-region--sheet">`)
	}
	for _, sep := range s.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range s.Label {
		renderLabel(sb, lab)
	}
	for _, g := range s.Group {
		renderGroup(ctx, sb, g, record, ro)
	}
	for _, f := range s.Field {
		renderField(ctx, sb, f, record, ro)
	}
	sb.WriteString(`</div>`)

	for _, nb := range s.Notebook {
		renderNotebook(ctx, sb, nb, record, ro, vr)
	}

	if hasSumTitle {
		sb.WriteString(`</div></div>`)
	}

	sb.WriteString(`</div>`)
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

func renderNotebook(ctx context.Context, sb *strings.Builder, nb parser.Notebook, record map[string]interface{}, ro bool, vr *ViewRecordData) {
	sb.WriteString(`<div class="sum-notebook sum-notebook--sheet">`)

	sb.WriteString(`<div class="sum-notebook-tabs" role="tablist">`)
	for i, p := range nb.Page {
		activeClass := "sum-notebook-tab"
		if i == 0 {
			activeClass += " sum-notebook-tab--active"
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" role="tab" class="%s">%s</button>`, activeClass, template.HTMLEscapeString(p.Title)))
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="sum-notebook-content">`)
	for i, p := range nb.Page {
		display := "none"
		if i == 0 {
			display = "block"
		}
		sb.WriteString(fmt.Sprintf(`<div class="sum-notebook-page sum-notebook-page--sheet" style="display: %s">`, display))

		pageTitle := strings.ToLower(strings.TrimSpace(p.Title))
		if vr != nil {
			if hooks, ok := NotebookHooks[vr.ResModel]; ok {
				if hook, ok := hooks[pageTitle]; ok {
					sb.WriteString(string(hook(ctx, vr, ro)))
					sb.WriteString(`</div>`)
					continue
				}
			}
		}

		if vr != nil && strings.TrimSpace(vr.ResModel) == "core.user" && pageTitle == "access rights" {
			writeResUsersSecuritySection(ctx, sb, vr, ro)
		} else {
			sb.WriteString(`<div class="sum-form-page-grid">`)
			for _, sep := range p.Separator {
				renderSeparator(sb, sep)
			}
			for _, lab := range p.Label {
				renderLabel(sb, lab)
			}
			for _, g := range p.Group {
				renderGroup(ctx, sb, g, record, ro)
			}
			for _, f := range p.Field {
				renderField(ctx, sb, f, record, ro)
			}
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
}

func renderSeparator(sb *strings.Builder, sep parser.Separator) {
	t := strings.TrimSpace(sep.String)
	if t != "" {
		sb.WriteString(`<div class="sum-separator sum-separator--title">` + template.HTMLEscapeString(t) + `</div>`)
	} else {
		sb.WriteString(`<hr class="sum-separator sum-separator--rule" />`)
	}
}

func renderLabel(sb *strings.Builder, lab parser.Label) {
	s := strings.TrimSpace(lab.String)
	if s == "" {
		return
	}
	id := strings.TrimSpace(lab.For)
	sb.WriteString(`<div class="sum-label sum-label--notebook">`)
	if id != "" {
		sb.WriteString(`<label class="sum-label-text" for="` + template.HTMLEscapeString(id) + `">` + template.HTMLEscapeString(s) + `</label>`)
	} else {
		sb.WriteString(`<span class="sum-label-text">` + template.HTMLEscapeString(s) + `</span>`)
	}
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

func renderGroup(ctx context.Context, sb *strings.Builder, g parser.Group, record map[string]interface{}, ro bool) {
	if ro {
		sb.WriteString(`<section class="sum-read-section">`)
		if g.Title != "" {
			sb.WriteString(`<h4 class="sum-read-section-title">` + template.HTMLEscapeString(strings.ToUpper(g.Title)) + `</h4>`)
		}
	} else {
		sb.WriteString(`<div class="sum-form-group">`)
		if g.Title != "" {
			sb.WriteString(`<h4 class="sum-form-group-title">` + template.HTMLEscapeString(g.Title) + `</h4>`)
		}
	}
	for _, sep := range g.Separator {
		renderSeparator(sb, sep)
	}
	for _, lab := range g.Label {
		renderLabel(sb, lab)
	}
	if ro {
		sb.WriteString(`<div class="sum-read-fields">`)
	} else {
		sb.WriteString(`<div class="sum-form-group-grid">`)
	}
	for _, f := range g.Field {
		renderField(ctx, sb, f, record, ro)
	}
	for _, subG := range g.Group {
		renderGroup(ctx, sb, subG, record, ro)
	}
	sb.WriteString(`</div>`)
	if ro {
		sb.WriteString(`</section>`)
	} else {
		sb.WriteString(`</div>`)
	}
}

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
