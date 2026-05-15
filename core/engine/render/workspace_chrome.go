package render

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"sumeru/core/engine/parser"
)

func renderWorkspaceFormToolbar(sb *strings.Builder, vr *ViewRecordData, header *parser.Header, record map[string]interface{}) {
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

	sb.WriteString(`<div class="sum-tree-control sum-ws-record-toolbar">`)
	sb.WriteString(`<div class="sum-tree-control-left">`)

	if hasID && !vr.FormEditing {
		sb.WriteString(`<a href="` + template.HTMLEscapeString(editURL) + `" class="sum-tree-btn-ghost">Edit</a>`)
	}
	if hasID && vr.FormEditing {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-tree-btn-new">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-tree-btn-ghost">Cancel</a>`)
	}
	if !hasID {
		sb.WriteString(`<button type="submit" form="sum-workspace-record-form" class="sum-tree-btn-new">Save</button>`)
		sb.WriteString(`<a href="` + template.HTMLEscapeString(cancelURL) + `" class="sum-tree-btn-ghost">Cancel</a>`)
	}

	if header != nil {
		for _, b := range header.Button {
			class := "sum-header-btn "
			if b.Class == "sum_highlight" {
				class += "sum-header-btn--primary"
			} else {
				class += "sum-header-btn--secondary"
			}
			sb.WriteString(fmt.Sprintf(`<button type="button" disabled class="%s sum-header-btn--disabled">%s</button>`, class, template.HTMLEscapeString(b.String)))
		}
	}

	sb.WriteString(`</div>`)

	sb.WriteString(`<div class="sum-ws-toolbar-right">`)
	if header != nil {
		writeStatusbarChips(sb, header, record)
	}
	sb.WriteString(`</div>`)

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
