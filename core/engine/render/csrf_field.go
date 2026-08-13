package render

import (
	"html/template"
	"strings"
)

func writeCSRFHidden(sb *strings.Builder, token string) {
	token = strings.TrimSpace(token)
	if token == "" {
		return
	}
	sb.WriteString(`<input type="hidden" name="csrf_token" value="` + template.HTMLEscapeString(token) + `" />`)
}
