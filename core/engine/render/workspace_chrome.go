package render

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

func renderWorkspaceFormChrome(sb *strings.Builder, vr *ViewRecordData) {
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
				qv.Set("view_type", "tree")
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

	sb.WriteString(`<div class="sum-ws-form-chrome">`)
	if hasID && !vr.FormEditing {
		sb.WriteString(`<a href="` + template.HTMLEscapeString(editURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Edit</a>`)
	}
	if hasID && vr.FormEditing {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-ws-btn sum-ws-btn--primary">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Cancel</a>`)
	}
	if !hasID {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-ws-btn sum-ws-btn--primary">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-ws-btn sum-ws-btn--ghost">Cancel</a>`)
	}
	sb.WriteString(`</div>`)

	sb.WriteString(`<form id="sum-workspace-record-form" method="post" action="` + template.HTMLEscapeString(saveAct) + `" class="sum-workspace-record-form">`)
	sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(vr.ResModel) + `" />`)
	if hasID {
		sb.WriteString(`<input type="hidden" name="id" value="` + template.HTMLEscapeString(fmt.Sprintf("%d", vr.RecordID)) + `" />`)
	}
	sb.WriteString(`<input type="hidden" name="next" value="` + nextEsc + `" />`)
}

func renderWorkspaceFormChromeClose(sb *strings.Builder, vr *ViewRecordData) {
	if workspaceFormChrome(vr) {
		sb.WriteString(`</form>`)
	}
}
