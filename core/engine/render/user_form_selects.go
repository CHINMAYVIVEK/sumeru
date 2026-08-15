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
		renderTypedInput(sb, f, label, record, ro, "text", fieldRequired("core.user", f.Name))
	}
}

// renderReadonlyFieldValue shows the same spacious field chrome as edit mode, without a dropdown.
func renderReadonlyFieldValue(sb *strings.Builder, f parser.Field, label, display string) {
	if strings.TrimSpace(display) == "" {
		display = "—"
	}
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" type="text" value="%s" readonly />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(display)))
	sb.WriteString(`</div>`)
}

// renderSelectShell renders an edit-mode <select>. Read mode must use renderReadonlyFieldValue instead.
func renderSelectShell(sb *strings.Builder, f parser.Field, label string, multi bool, body func()) {
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	multiAttr := ""
	if multi {
		multiAttr = ` multiple`
	}
	sb.WriteString(fmt.Sprintf(`<select class="sum-field-input sum-field-select" id="%s" name="%s"%s>`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), multiAttr))
	body()
	sb.WriteString(`</select></div>`)
}

func writeOption(sb *strings.Builder, value, label string, selected bool) {
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
	if ro {
		display := "Inactive"
		if active {
			display = "Active"
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		writeOption(sb, "true", "Active", active)
		writeOption(sb, "false", "Inactive", !active)
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
	if ro {
		display := cur
		for _, o := range opts {
			if o[0] == cur {
				display = o[1]
				break
			}
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		for _, o := range opts {
			writeOption(sb, o[0], o[1], o[0] == cur)
		}
	})
}

func collectCompanyIDs(record map[string]interface{}, multi bool) map[int]struct{} {
	selected := map[int]struct{}{}
	if !multi {
		if n, ok := orm.CoerceInt64(record["company_id"]); ok && n > 0 {
			selected[int(n)] = struct{}{}
		}
		return selected
	}
	raw, ok := record["company_ids"]
	if !ok || raw == nil {
		if s := strings.TrimSpace(recStr(record, "company_ids")); s != "" {
			for _, part := range strings.Split(s, ",") {
				if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n > 0 {
					selected[n] = struct{}{}
				}
			}
		}
		return selected
	}
	switch ids := raw.(type) {
	case []int:
		for _, id := range ids {
			if id > 0 {
				selected[id] = struct{}{}
			}
		}
	case []int64:
		for _, id := range ids {
			if id > 0 {
				selected[int(id)] = struct{}{}
			}
		}
	case []interface{}:
		for _, v := range ids {
			if n, ok := orm.CoerceInt64(v); ok && n > 0 {
				selected[int(n)] = struct{}{}
			}
		}
	default:
		if s := strings.TrimSpace(orm.AsString(raw)); s != "" {
			for _, part := range strings.Split(s, ",") {
				if n, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && n > 0 {
					selected[n] = struct{}{}
				}
			}
		}
	}
	return selected
}

func renderCompanySelect(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro, multi bool) {
	companies, err := orm.Search(ctx, "core.company", nil)
	if err != nil {
		companies = nil
	}
	selected := collectCompanyIDs(record, multi)

	var selectedIDs, availableIDs []int
	var selectedNames, availableNames []string
	for _, c := range companies {
		id, ok := orm.CoerceInt64(c["id"])
		if !ok || id <= 0 {
			continue
		}
		nm := strings.TrimSpace(orm.AsString(c["name"]))
		if nm == "" {
			nm = fmt.Sprintf("#%d", int(id))
		}
		if _, ok := selected[int(id)]; ok {
			selectedIDs = append(selectedIDs, int(id))
			selectedNames = append(selectedNames, nm)
		} else {
			availableIDs = append(availableIDs, int(id))
			availableNames = append(availableNames, nm)
		}
	}

	if ro {
		if multi {
			renderReadonlyCompanyTags(sb, f, label, selectedNames)
		} else {
			renderReadonlyFieldValue(sb, f, label, strings.Join(selectedNames, ", "))
		}
		return
	}

	if multi {
		renderCompanyMultiSelect(sb, f, label, selectedIDs, selectedNames, availableIDs, availableNames)
		return
	}

	renderSelectShell(sb, f, label, false, func() {
		writeOption(sb, "", "—", len(selected) == 0)
		for i := range selectedIDs {
			writeOption(sb, strconv.Itoa(selectedIDs[i]), selectedNames[i], true)
		}
		for i := range availableIDs {
			writeOption(sb, strconv.Itoa(availableIDs[i]), availableNames[i], false)
		}
	})
}

func renderReadonlyCompanyTags(sb *strings.Builder, f parser.Field, label string, names []string) {
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	if len(names) == 0 {
		sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" type="text" value="—" readonly />`,
			template.HTMLEscapeString(f.Name)))
		sb.WriteString(`</div>`)
		return
	}
	sb.WriteString(`<div class="sum-multi-select-tags sum-multi-select-tags--readonly" id="` + template.HTMLEscapeString(f.Name) + `">`)
	for _, nm := range names {
		sb.WriteString(`<span class="sum-multi-select-tag">` + template.HTMLEscapeString(nm) + `</span>`)
	}
	sb.WriteString(`</div></div>`)
}

func renderCompanyMultiSelect(sb *strings.Builder, f parser.Field, label string, selectedIDs []int, selectedNames []string, availableIDs []int, availableNames []string) {
	name := template.HTMLEscapeString(f.Name)
	sb.WriteString(`<div class="sum-field-widget sum-multi-select" data-sum-multi-select data-name="` + name + `">`)
	sb.WriteString(`<label class="sum-field-label" for="` + name + `_add">` + template.HTMLEscapeString(label) + `</label>`)
	// Sentinel so clearing all companies still posts company_ids and updates links.
	sb.WriteString(`<input type="hidden" name="` + name + `" value="" />`)
	sb.WriteString(`<div class="sum-multi-select-box">`)
	sb.WriteString(`<select class="sum-field-input sum-field-select sum-multi-select-add" id="` + name + `_add" data-sum-multi-add>`)
	sb.WriteString(`<option value="">Add company…</option>`)
	for i := range availableIDs {
		sb.WriteString(fmt.Sprintf(`<option value="%d">%s</option>`, availableIDs[i], template.HTMLEscapeString(availableNames[i])))
	}
	sb.WriteString(`</select>`)
	sb.WriteString(`<div class="sum-multi-select-tags" data-sum-multi-tags>`)
	for i := range selectedIDs {
		id := selectedIDs[i]
		nm := selectedNames[i]
		sb.WriteString(`<span class="sum-multi-select-tag" data-id="` + strconv.Itoa(id) + `">`)
		sb.WriteString(`<span class="sum-multi-select-tag-label">` + template.HTMLEscapeString(nm) + `</span>`)
		sb.WriteString(`<button type="button" class="sum-multi-select-remove" data-sum-multi-remove aria-label="Remove">×</button>`)
		sb.WriteString(`<input type="hidden" name="` + name + `" value="` + strconv.Itoa(id) + `" />`)
		sb.WriteString(`</span>`)
	}
	sb.WriteString(`</div></div></div>`)
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
	if ro {
		display := cur
		for _, l := range langs {
			if orm.AsString(l["code"]) == cur {
				nm := orm.AsString(l["name"])
				if nm != "" {
					display = nm
				}
				break
			}
		}
		if len(langs) == 0 && cur == "en_US" {
			display = "English (US)"
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		if len(langs) == 0 {
			writeOption(sb, "en_US", "English (US)", cur == "en_US")
			return
		}
		for _, l := range langs {
			code := orm.AsString(l["code"])
			nm := orm.AsString(l["name"])
			if nm == "" {
				nm = code
			}
			writeOption(sb, code, nm, code == cur)
		}
	})
}

func renderTimezoneSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	if ro {
		display := cur
		if display == "" {
			display = "—"
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		writeOption(sb, "", "—", cur == "")
		found := false
		for _, tz := range commonUserTimezones {
			sel := tz.Value == cur
			if sel {
				found = true
			}
			writeOption(sb, tz.Value, tz.Label, sel)
		}
		if cur != "" && !found {
			writeOption(sb, cur, cur, true)
		}
	})
}
