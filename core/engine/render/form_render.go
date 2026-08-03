package render

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
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
