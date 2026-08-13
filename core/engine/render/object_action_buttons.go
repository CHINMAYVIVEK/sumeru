package render

import (
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
)

func writeObjectActionButtons(sb *strings.Builder, header *parser.Header, vr *ViewRecordData, nextURL string) {
	if header == nil || vr == nil || vr.RecordID <= 0 {
		return
	}
	for _, b := range header.Button {
		writeObjectActionButton(sb, b, vr.ResModel, vr.RecordID, nextURL, vr.CSRFToken, false)
	}
}

func writeObjectActionButton(sb *strings.Builder, b parser.Button, model string, id int, nextURL, csrfToken string, insideForm bool) {
	label := strings.TrimSpace(b.String)
	if label == "" {
		label = strings.TrimSpace(b.Name)
	}
	class := "sum-header-btn sum-header-btn--secondary"
	if strings.Contains(b.Class, "sum_highlight") || strings.Contains(b.Class, "btn-primary") {
		class = "sum-header-btn sum-header-btn--primary"
	}
	if strings.Contains(b.Class, "sum-form-footer-btn") || insideForm {
		class = "sum-form-footer-btn"
		if strings.Contains(b.Class, "sum_highlight") || strings.Contains(b.Class, "btn-primary") {
			class += " sum-form-footer-btn--primary"
		}
	}
	btnType := strings.ToLower(strings.TrimSpace(b.Type))
	method := strings.TrimSpace(b.Name)
	if (btnType == "object" || btnType == "") && method != "" && model != "" && id > 0 {
		if insideForm {
			// Submit parent record form to object action so edited fields are included.
			sb.WriteString(fmt.Sprintf(
				`<button type="submit" formaction="/web/action/object" formmethod="post" name="method" value="%s" class="%s">%s</button>`,
				template.HTMLEscapeString(method), class, template.HTMLEscapeString(label),
			))
			return
		}
		formID := fmt.Sprintf("sum-obj-act-%s-%d", sanitizeFormID(method), id)
		sb.WriteString(fmt.Sprintf(
			`<form id="%s" method="post" action="/web/action/object" class="sum-object-action-form" style="display:inline">`,
			template.HTMLEscapeString(formID),
		))
		sb.WriteString(`<input type="hidden" name="model" value="` + template.HTMLEscapeString(model) + `" />`)
		sb.WriteString(fmt.Sprintf(`<input type="hidden" name="id" value="%d" />`, id))
		sb.WriteString(`<input type="hidden" name="method" value="` + template.HTMLEscapeString(method) + `" />`)
		if nextURL != "" {
			sb.WriteString(`<input type="hidden" name="next" value="` + template.HTMLEscapeString(nextURL) + `" />`)
		}
		writeCSRFHidden(sb, csrfToken)
		sb.WriteString(fmt.Sprintf(`<button type="submit" class="%s">%s</button>`, class, template.HTMLEscapeString(label)))
		sb.WriteString(`</form>`)
		return
	}
	sb.WriteString(fmt.Sprintf(`<button type="button" disabled class="%s sum-header-btn--disabled" title="Action unavailable">%s</button>`,
		class, template.HTMLEscapeString(label)))
}

func sanitizeFormID(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	out := b.String()
	if out == "" {
		return "act"
	}
	return out
}
