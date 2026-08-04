package render

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// commonUserTimezones is the hardcoded IANA list for core.user tz dropdown.
var commonUserTimezones = []struct{ Value, Label string }{
	{"UTC", "UTC"},
	{"Asia/Kolkata", "Asia/Kolkata"},
	{"Asia/Dubai", "Asia/Dubai"},
	{"Asia/Singapore", "Asia/Singapore"},
	{"Asia/Tokyo", "Asia/Tokyo"},
	{"Asia/Shanghai", "Asia/Shanghai"},
	{"Europe/London", "Europe/London"},
	{"Europe/Paris", "Europe/Paris"},
	{"Europe/Berlin", "Europe/Berlin"},
	{"Europe/Amsterdam", "Europe/Amsterdam"},
	{"America/New_York", "America/New_York"},
	{"America/Chicago", "America/Chicago"},
	{"America/Denver", "America/Denver"},
	{"America/Los_Angeles", "America/Los_Angeles"},
	{"America/Sao_Paulo", "America/Sao_Paulo"},
	{"Australia/Sydney", "Australia/Sydney"},
	{"Pacific/Auckland", "Pacific/Auckland"},
}

func isCoreUserSelectField(resModel, fieldName string, widget string) bool {
	if resModel != "core.user" {
		return false
	}
	switch fieldName {
	case "active", "user_type", "company_id", "company_ids", "lang", "tz":
		return true
	}
	return widget == "selection" && (fieldName == "active" || fieldName == "user_type" || fieldName == "lang" || fieldName == "tz")
}

func renderCoreUserSelectField(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	switch f.Name {
	case "active":
		renderActiveSelect(sb, f, label, record, ro)
	case "user_type":
		renderUserTypeSelect(sb, f, label, record, ro)
	case "company_id":
		renderCompanySelect(ctx, sb, f, label, record, ro, false)
	case "company_ids":
		renderCompanySelect(ctx, sb, f, label, record, ro, true)
	case "lang":
		renderLangSelect(ctx, sb, f, label, record, ro)
	case "tz":
		renderTimezoneSelect(sb, f, label, record, ro)
	default:
		renderTypedInput(sb, f, label, record, ro, "text")
	}
}

func renderSelectShell(sb *strings.Builder, f parser.Field, label string, ro bool, multi bool, body func()) {
	if ro {
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value">`)
		body()
		sb.WriteString(`</div></div>`)
		return
	}
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	multiAttr := ""
	if multi {
		multiAttr = ` multiple size="4"`
	}
	sb.WriteString(fmt.Sprintf(`<select class="sum-field-input sum-field-select" id="%s" name="%s"%s>`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), multiAttr))
	body()
	sb.WriteString(`</select></div>`)
}

func writeOption(sb *strings.Builder, value, label string, selected bool, ro bool) {
	if ro {
		if selected {
			sb.WriteString(template.HTMLEscapeString(label))
		}
		return
	}
	sel := ""
	if selected {
		sel = ` selected`
	}
	sb.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`,
		template.HTMLEscapeString(value), sel, template.HTMLEscapeString(label)))
}

func renderActiveSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	raw, _ := rawField(record, f.Name)
	active := true
	if raw != nil {
		active = isTruthyDB(raw)
	}
	renderSelectShell(sb, f, label, ro, false, func() {
		if ro {
			if active {
				sb.WriteString("Active")
			} else {
				sb.WriteString("Inactive")
			}
			return
		}
		writeOption(sb, "true", "Active", active, false)
		writeOption(sb, "false", "Inactive", !active, false)
	})
}

func renderUserTypeSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	if cur == "" {
		cur = "internal"
	}
	opts := [][]string{{"internal", "Internal"}, {"portal", "Portal"}, {"public", "Public"}}
	if fd := fieldDef("core.user", "user_type"); fd != nil && len(fd.Selection) > 0 {
		opts = fd.Selection
	}
	renderSelectShell(sb, f, label, ro, false, func() {
		if ro {
			for _, o := range opts {
				if o[0] == cur {
					sb.WriteString(template.HTMLEscapeString(o[1]))
					return
				}
			}
			sb.WriteString(template.HTMLEscapeString(cur))
			return
		}
		for _, o := range opts {
			writeOption(sb, o[0], o[1], o[0] == cur, false)
		}
	})
}

func renderCompanySelect(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro, multi bool) {
	companies, err := orm.Search(ctx, "core.company", nil)
	if err != nil {
		companies = nil
	}
	selected := map[int]struct{}{}
	if multi {
		if ids, ok := record["company_ids"].([]int); ok {
			for _, id := range ids {
				selected[id] = struct{}{}
			}
		} else if s := strings.TrimSpace(recStr(record, "company_ids")); s != "" {
			for _, part := range strings.Split(s, ",") {
				if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n > 0 {
					selected[n] = struct{}{}
				}
			}
		}
	} else if n, ok := orm.CoerceInt64(record["company_id"]); ok && n > 0 {
		selected[int(n)] = struct{}{}
	}

	renderSelectShell(sb, f, label, ro, multi, func() {
		if ro {
			var names []string
			for _, c := range companies {
				id, _ := orm.CoerceInt64(c["id"])
				if _, ok := selected[int(id)]; ok {
					names = append(names, orm.AsString(c["name"]))
				}
			}
			if len(names) == 0 {
				sb.WriteString("—")
			} else {
				sb.WriteString(template.HTMLEscapeString(strings.Join(names, ", ")))
			}
			return
		}
		if !multi {
			writeOption(sb, "", "—", len(selected) == 0, false)
		}
		for _, c := range companies {
			id, ok := orm.CoerceInt64(c["id"])
			if !ok || id <= 0 {
				continue
			}
			nm := orm.AsString(c["name"])
			_, sel := selected[int(id)]
			writeOption(sb, fmt.Sprintf("%d", int(id)), nm, sel, false)
		}
	})
}

func renderLangSelect(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	if cur == "" {
		cur = "en_US"
	}
	allLangs, err := orm.Search(ctx, "core.lang", nil)
	var langs []map[string]interface{}
	if err == nil {
		for _, l := range allLangs {
			if isTruthyDB(l["active"]) || l["active"] == nil {
				langs = append(langs, l)
			}
		}
	}
	renderSelectShell(sb, f, label, ro, false, func() {
		if ro {
			for _, l := range langs {
				if orm.AsString(l["code"]) == cur {
					sb.WriteString(template.HTMLEscapeString(orm.AsString(l["name"])))
					return
				}
			}
			sb.WriteString(template.HTMLEscapeString(cur))
			return
		}
		if len(langs) == 0 {
			writeOption(sb, "en_US", "English (US)", cur == "en_US", false)
			return
		}
		for _, l := range langs {
			code := orm.AsString(l["code"])
			nm := orm.AsString(l["name"])
			if nm == "" {
				nm = code
			}
			writeOption(sb, code, nm, code == cur, false)
		}
	})
}

func renderTimezoneSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	renderSelectShell(sb, f, label, ro, false, func() {
		if ro {
			if cur == "" {
				sb.WriteString("—")
			} else {
				sb.WriteString(template.HTMLEscapeString(cur))
			}
			return
		}
		writeOption(sb, "", "—", cur == "", false)
		found := false
		for _, tz := range commonUserTimezones {
			sel := tz.Value == cur
			if sel {
				found = true
			}
			writeOption(sb, tz.Value, tz.Label, sel, false)
		}
		if cur != "" && !found {
			writeOption(sb, cur, cur, true, false)
		}
	})
}
