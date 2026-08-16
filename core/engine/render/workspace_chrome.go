package render

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

func renderWorkspaceFormToolbar(ctx context.Context, sb *strings.Builder, vr *ViewRecordData, view *parser.View, header *parser.Header, record map[string]interface{}) {
	if vr == nil || strings.TrimSpace(vr.ResModel) == "" {
		return
	}
	base := strings.TrimSpace(vr.FormBaseQuery)
	hasID := vr.RecordID > 0
	editQS := "edit=1"
	if base != "" {
		editQS = base + "&edit=1"
	}
	editURL := "/web?" + editQS
	cancelURL := "/web"
	if base != "" {
		if !hasID {
			if qv, err := url.ParseQuery(base); err == nil {
				qv.Set("view_type", "list")
				qv.Del("id")
				cancelURL = "/web?" + qv.Encode()
			} else {
				cancelURL = "/web?" + base
			}
		} else {
			cancelURL = "/web?" + base
		}
	}
	saveAct := strings.TrimSpace(vr.FormSaveAction)
	if saveAct == "" {
		saveAct = "/web/record/save"
	}
	nextEsc := template.HTMLEscapeString(cancelURL)

	sb.WriteString(`<div class="sum-list-control sum-ws-record-toolbar">`)
	sb.WriteString(`<div class="sum-list-control-left">`)

	if hasID && !vr.FormEditing {
		sb.WriteString(`<a href="` + template.HTMLEscapeString(editURL) + `" class="sum-list-btn-ghost">Edit</a>`)
	}
	if !hasID || vr.FormEditing {
		writeSaveCancelButtons(sb, cancelURL)
	}

	if view != nil {
		caps := report.CapabilitiesFromView(view)
		recordID := 0
		if hasID {
			recordID = vr.RecordID
		}
		sb.WriteString(RenderReportToolbar(caps, view.Model, vr.ActionID, menuIDFromFormBaseQuery(vr.FormBaseQuery), recordID, viewFieldsForReport(view), vr.CSRFToken))
	}

	if header != nil {
		writeObjectActionButtons(sb, header, vr, cancelURL)
	}

	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="sum-ws-toolbar-right">`)
	if header != nil {
		writeStatusbarChips(ctx, sb, header, record, vr)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)

	sb.WriteString(`<form id="sum-workspace-record-form" method="post" action="` + template.HTMLEscapeString(saveAct) + `" class="sum-workspace-record-form">`)
	sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(vr.ResModel) + `" />`)
	if hasID {
		sb.WriteString(`<input type="hidden" name="id" value="` + template.HTMLEscapeString(fmt.Sprintf("%d", vr.RecordID)) + `" />`)
	}
	sb.WriteString(`<input type="hidden" name="next" value="` + nextEsc + `" />`)
	writeCSRFHidden(sb, vr.CSRFToken)
}

func renderWorkspaceFormChromeClose(sb *strings.Builder, vr *ViewRecordData) {
	if workspaceFormChrome(vr) {
		sb.WriteString(`</form>`)
	}
}
