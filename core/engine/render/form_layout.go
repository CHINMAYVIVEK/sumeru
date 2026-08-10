package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
)

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

	var titleDiv *parser.Div
	for i := range s.Div {
		if strings.Contains(s.Div[i].Class, "sum_title") {
			titleDiv = &s.Div[i]
			break
		}
	}

	if hasSumTitle && titleDiv != nil {
		sb.WriteString(`<div class="sum-form-split-layout sum-form-split-layout--compact" data-sum-form-split>`)
		sb.WriteString(`<aside class="sum-form-split-left sum-form-split-left--avatar" aria-label="Profile picture">`)
		renderTitleAvatar(ctx, sb, *titleDiv, record, ro)
		sb.WriteString(`</aside>`)
		sb.WriteString(`<div class="sum-form-split-main">`)
		renderTitleBody(ctx, sb, *titleDiv, record, ro)
	} else if titleDiv != nil {
		renderTitle(ctx, sb, *titleDiv, record, ro)
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
		renderGroup(ctx, sb, g, record, ro, vr)
	}
	for _, f := range s.Field {
		renderField(ctx, sb, f, record, ro, vr)
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
				renderGroup(ctx, sb, g, record, ro, vr)
			}
			for _, f := range p.Field {
				renderField(ctx, sb, f, record, ro, vr)
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

func renderGroup(ctx context.Context, sb *strings.Builder, g parser.Group, record map[string]interface{}, ro bool, vr *ViewRecordData) {
	untitled := strings.TrimSpace(g.Title) == ""
	if ro {
		cls := "sum-read-section sum-read-section--full"
		if untitled {
			cls += " sum-read-section--plain"
		}
		sb.WriteString(`<section class="` + cls + `">`)
		if g.Title != "" {
			sb.WriteString(`<h4 class="sum-read-section-title">` + template.HTMLEscapeString(strings.ToUpper(g.Title)) + `</h4>`)
		}
	} else {
		cls := "sum-form-group sum-form-group--full"
		if untitled {
			cls += " sum-form-group--plain"
		}
		sb.WriteString(`<div class="` + cls + `">`)
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
		renderField(ctx, sb, f, record, ro, vr)
	}
	for _, subG := range g.Group {
		renderGroup(ctx, sb, subG, record, ro, vr)
	}
	sb.WriteString(`</div>`)
	if ro {
		sb.WriteString(`</section>`)
	} else {
		sb.WriteString(`</div>`)
	}
}
