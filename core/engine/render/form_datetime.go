package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func fieldInputAttrs(label string, required bool) string {
	if !required {
		return ""
	}
	return ` data-required="1" aria-required="true" data-field-label="` + template.HTMLEscapeString(label) + `"`
}

func resolveDateInputType(f parser.Field, fd *orm.FieldDefinition) string {
	switch strings.ToLower(strings.TrimSpace(f.Widget)) {
	case "date":
		return "date"
	case "datetime":
		return "datetime-local"
	case "time":
		return "time"
	}
	if mode := fieldOptionMode(f.Options); mode != "" {
		switch mode {
		case "date":
			return "date"
		case "datetime":
			return "datetime-local"
		case "time":
			return "time"
		}
	}
	if fd != nil {
		switch fd.Type {
		case orm.Date:
			return "date"
		case orm.DateTime:
			return "datetime-local"
		}
	}
	return "text"
}

func fieldOptionMode(optionsRaw string) string {
	optionsRaw = strings.TrimSpace(optionsRaw)
	if optionsRaw == "" {
		return ""
	}
	var opts map[string]interface{}
	if err := json.Unmarshal([]byte(optionsRaw), &opts); err != nil {
		return ""
	}
	mode, _ := opts["mode"].(string)
	return strings.ToLower(strings.TrimSpace(mode))
}

func formatDateInputValue(raw, inputType string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	switch inputType {
	case "date":
		if len(s) >= 10 {
			return s[:10]
		}
	case "datetime-local":
		s = strings.Replace(s, " ", "T", 1)
		if len(s) >= 16 {
			return s[:16]
		}
	case "time":
		if len(s) >= 5 {
			return s[:5]
		}
	}
	return s
}

func renderDateTimeField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, inputType string, required bool) {
	val := formatDateInputValue(recStr(record, f.Name), inputType)
	if ro {
		renderReadonlyFieldValue(sb, f, label, val)
		return
	}
	attrs := fieldInputAttrs(label, required)
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" name="%s" type="%s" value="%s"%s />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(inputType),
		template.HTMLEscapeString(val), attrs))
	sb.WriteString(`</div>`)
}
