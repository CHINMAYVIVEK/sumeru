package render

import (
	"net/url"
	"strings"

	"sumeru/core/engine/parser"
)

func menuIDFromFormBaseQuery(qs string) string {
	qs = strings.TrimSpace(qs)
	if qs == "" {
		return ""
	}
	qv, err := url.ParseQuery(qs)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(qv.Get("menu_id"))
}

func viewFieldsForReport(view *parser.View) []parser.Field {
	if view == nil {
		return nil
	}
	if len(view.Field) > 0 {
		return view.Field
	}
	var out []parser.Field
	if view.Sheet != nil {
		for _, f := range view.Sheet.Field {
			out = append(out, f)
		}
	}
	return out
}
