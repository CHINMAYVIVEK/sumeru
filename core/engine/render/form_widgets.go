package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func renderField(ctx context.Context, sb *strings.Builder, f parser.Field, record map[string]interface{}, ro bool, vr *ViewRecordData) {
	if gs := strings.TrimSpace(f.Groups); gs != "" {
		if !orm.UserHasAnyAccessGroup(ctx, orm.SecurityUID(ctx), gs) {
			return
		}
	}
	label := f.Label
	if label == "" {
		label = strings.Title(strings.ReplaceAll(f.Name, "_", " "))
	}

	resModel := ""
	if vr != nil {
		resModel = strings.TrimSpace(vr.ResModel)
	}
	if isCoreUserSelectField(resModel, f.Name, f.Widget) {
		renderCoreUserSelectField(ctx, sb, f, label, record, ro)
		return
	}
	if fd := fieldDef(resModel, f.Name); fd != nil {
		switch fd.Type {
		case orm.Many2One:
			if strings.EqualFold(f.Widget, "selection") || strings.EqualFold(f.Widget, "dropdown") {
				renderMany2OneSelectField(ctx, sb, f, label, record, ro, fd.Relation, resModel)
				return
			}
			renderMany2OneField(ctx, sb, f, label, record, ro, fd.Relation)
			return
		case orm.One2Many:
			renderOne2ManyField(ctx, sb, f, label, record, ro, resModel, fd.Relation, vr)
			return
		case orm.Text:
			if f.Widget == "image" || f.Name == "image" {
				renderImageField(sb, f, label, record, ro)
				return
			}
			renderTextareaField(sb, f, label, record, ro)
			return
		case orm.Boolean:
			renderBooleanField(sb, f, label, record, ro)
			return
		case orm.Selection:
			if len(fd.Selection) > 0 {
				renderModelSelectionSelect(sb, f, label, record, ro, fd.Selection)
				return
			}
		}
	}

	isBoolish := strings.EqualFold(f.Widget, "boolean") ||
		strings.EqualFold(f.Widget, "radio") ||
		strings.HasSuffix(f.Name, "_active")
	if isBoolish {
		renderBooleanField(sb, f, label, record, ro)
		return
	}

	if f.Widget == "image" {
		renderImageField(sb, f, label, record, ro)
		return
	}
	if f.Widget == "many2many_tags" {
		txt := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(txt) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(`<div class="sum-field-tags">`)
		sb.WriteString(template.HTMLEscapeString(txt))
		sb.WriteString(`</div></div>`)
		return
	}

	if f.Widget == "email" || f.Widget == "phone" || f.Widget == "tel" ||
		strings.Contains(strings.ToLower(f.Name), "email") || strings.Contains(strings.ToLower(f.Name), "phone") {
		inputType := "text"
		switch {
		case f.Widget == "email" || strings.Contains(strings.ToLower(f.Name), "email"):
			inputType = "email"
		case f.Widget == "phone" || f.Widget == "tel" || strings.Contains(strings.ToLower(f.Name), "phone"):
			inputType = "tel"
		}
		renderTypedInput(sb, f, label, record, ro, inputType)
		return
	}

	if f.Widget == "text" || strings.Contains(strings.ToLower(f.Name), "note") || strings.Contains(strings.ToLower(f.Name), "comment") {
		renderTextareaField(sb, f, label, record, ro)
		return
	}

	renderTypedInput(sb, f, label, record, ro, "text")
}

func renderBooleanField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	raw, hasRaw := rawField(record, f.Name)
	truthy := hasRaw && isTruthyDB(raw)
	if ro {
		val := "No"
		if truthy {
			val = "Yes"
		}
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value sum-read-value--bool">` + template.HTMLEscapeString(val) + `</div>`)
		sb.WriteString(`</div>`)
		return
	}
	if strings.EqualFold(f.Widget, "radio") {
		renderBooleanRadio(sb, f, label, truthy)
		return
	}
	renderBooleanSelect(sb, f, label, truthy)
}

func renderBooleanSelect(sb *strings.Builder, f parser.Field, label string, truthy bool) {
	trueLabel, falseLabel := "Yes", "No"
	if f.Name == "active" {
		trueLabel, falseLabel = "Active", "Inactive"
	}
	renderSelectShell(sb, f, label, false, false, func() {
		writeOption(sb, "true", trueLabel, truthy, false)
		writeOption(sb, "false", falseLabel, !truthy, false)
	})
}

func renderBooleanRadio(sb *strings.Builder, f parser.Field, label string, truthy bool) {
	name := template.HTMLEscapeString(f.Name)
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<span class="sum-field-label">` + template.HTMLEscapeString(label) + `</span>`)
	sb.WriteString(`<div class="sum-field-radio-group" role="radiogroup" aria-label="` + template.HTMLEscapeString(label) + `">`)
	yesChecked, noChecked := "", ""
	if truthy {
		yesChecked = ` checked`
	} else {
		noChecked = ` checked`
	}
	sb.WriteString(fmt.Sprintf(`<label class="sum-field-radio"><input type="radio" name="%s" value="true"%s /><span>Yes</span></label>`, name, yesChecked))
	sb.WriteString(fmt.Sprintf(`<label class="sum-field-radio"><input type="radio" name="%s" value="false"%s /><span>No</span></label>`, name, noChecked))
	sb.WriteString(`</div></div>`)
}

func renderModelSelectionSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, opts [][]string) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	renderSelectShell(sb, f, label, ro, false, func() {
		if ro {
			for _, o := range opts {
				if len(o) >= 2 && o[0] == cur {
					sb.WriteString(template.HTMLEscapeString(o[1]))
					return
				}
			}
			if cur == "" {
				sb.WriteString("—")
			} else {
				sb.WriteString(template.HTMLEscapeString(cur))
			}
			return
		}
		writeOption(sb, "", "—", cur == "", false)
		for _, o := range opts {
			if len(o) < 2 {
				continue
			}
			writeOption(sb, o[0], o[1], o[0] == cur, false)
		}
	})
}

func imageSrcOK(src string) bool {
	return src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:"))
}

// renderImageField renders widget="image" with upload controls (data-URL via JS).
func renderImageField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	src := strings.TrimSpace(recStr(record, f.Name))
	if ro {
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value">`)
		if imageSrcOK(src) {
			sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-read-image" />`, template.HTMLEscapeString(src)))
		} else {
			sb.WriteString(`<span class="sum-read-empty">—</span>`)
		}
		sb.WriteString(`</div></div>`)
		return
	}
	sb.WriteString(`<div class="sum-field-widget" data-sum-image>`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	if imageSrcOK(src) {
		sb.WriteString(`<div class="sum-image-thumb">`)
		sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-image-thumb-img" data-sum-image-preview />`, template.HTMLEscapeString(src)))
		sb.WriteString(`</div>`)
	} else {
		sb.WriteString(`<div class="sum-image-placeholder" data-sum-image-placeholder>`)
		sb.WriteString(`<svg class="sum-image-placeholder-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)
		sb.WriteString(`</div>`)
		sb.WriteString(`<div class="sum-image-thumb" hidden>`)
		sb.WriteString(`<img src="" alt="" class="sum-image-thumb-img" data-sum-image-preview hidden />`)
		sb.WriteString(`</div>`)
	}
	sb.WriteString(fmt.Sprintf(`<input type="hidden" id="%s" name="%s" value="%s" data-sum-image-value />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(src)))
	sb.WriteString(`<label class="sum-form-avatar-upload sum-image-upload">`)
	sb.WriteString(`<input type="file" accept="image/*" data-sum-image-file />`)
	sb.WriteString(`<span>Change</span>`)
	sb.WriteString(`</label>`)
	sb.WriteString(`</div>`)
}

func fieldDef(modelName, fieldName string) *orm.FieldDefinition {
	if modelName == "" || fieldName == "" {
		return nil
	}
	m, ok := orm.Registry[modelName]
	if !ok || m == nil {
		return nil
	}
	for _, f := range m.Fields() {
		if f.Name == fieldName {
			fd := f
			return &fd
		}
	}
	return nil
}

func renderTypedInput(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, inputType string) {
	placeholder := f.Placeholder
	if placeholder == "" {
		placeholder = "Enter " + strings.ToLower(label) + "..."
	}
	val := recStr(record, f.Name)
	if ro {
		display := val
		if display == "" {
			display = "—"
		}
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(display) + `</div>`)
		sb.WriteString(`</div>`)
		return
	}
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" name="%s" type="%s" placeholder="%s" value="%s" />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(inputType),
		template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val)))
	sb.WriteString(`</div>`)
}

func renderTextareaField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	val := recStr(record, f.Name)
	if ro {
		display := val
		if display == "" {
			display = "—"
		}
		sb.WriteString(`<div class="sum-read-field sum-read-field--row sum-read-field--full">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value sum-read-value--pre">` + template.HTMLEscapeString(display) + `</div>`)
		sb.WriteString(`</div>`)
		return
	}
	placeholder := f.Placeholder
	if placeholder == "" {
		placeholder = "Enter " + strings.ToLower(label) + "..."
	}
	sb.WriteString(`<div class="sum-field-widget sum-field-widget--full">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<textarea class="sum-field-input sum-field-textarea" id="%s" name="%s" rows="4" placeholder="%s">%s</textarea>`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val)))
	sb.WriteString(`</div>`)
}

func renderMany2OneField(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, comodel string) {
	id := 0
	if raw, ok := rawField(record, f.Name); ok {
		if n, ok := orm.CoerceInt64(raw); ok {
			id = int(n)
		}
	}
	display := ""
	if id > 0 && comodel != "" {
		display = orm.DisplayNameForID(ctx, comodel, id)
	}
	if ro {
		if display == "" {
			display = "—"
		}
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(display) + `</div>`)
		sb.WriteString(`</div>`)
		return
	}
	valAttr := ""
	if id > 0 {
		valAttr = fmt.Sprintf("%d", id)
	}
	sb.WriteString(`<div class="sum-field-widget sum-m2o" data-sum-m2o data-comodel="` + template.HTMLEscapeString(comodel) + `">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `_search">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" id="%s" value="%s" data-sum-m2o-id />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(valAttr)))
	sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s_search" type="search" autocomplete="off" placeholder="Search…" value="%s" data-sum-m2o-search />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(display)))
	sb.WriteString(`<ul class="sum-m2o-results" data-sum-m2o-results hidden></ul>`)
	sb.WriteString(`</div>`)
}

// cascadeParentForField returns the parent field name used to filter dropdown options.
func cascadeParentForField(fieldName string) (parent string, fallback string) {
	switch fieldName {
	case "state_id":
		return "country_id", ""
	case "city_id":
		return "state_id", "country_id"
	default:
		return "", ""
	}
}

func renderMany2OneSelectField(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, comodel, resModel string) {
	id := 0
	if raw, ok := rawField(record, f.Name); ok {
		if n, ok := orm.CoerceInt64(raw); ok {
			id = int(n)
		}
	}
	display := ""
	if id > 0 && comodel != "" {
		display = orm.DisplayNameForID(ctx, comodel, id)
	}
	if ro {
		if display == "" {
			display = "—"
		}
		sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
		sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
		sb.WriteString(`<div class="sum-read-value">` + template.HTMLEscapeString(display) + `</div>`)
		sb.WriteString(`</div>`)
		return
	}

	parentField, fallbackParent := cascadeParentForField(f.Name)
	var filterField string
	var filterID int64
	if parentField != "" {
		if pid, ok := orm.CoerceInt64(record[parentField]); ok && pid > 0 {
			filterField, filterID = parentField, pid
		} else if fallbackParent != "" {
			if pid, ok := orm.CoerceInt64(record[fallbackParent]); ok && pid > 0 {
				filterField, filterID = fallbackParent, pid
			} else {
				filterField, filterID = parentField, 0
			}
		} else {
			filterField, filterID = parentField, 0
		}
	}

	rows, err := orm.RelNameSearchFiltered(ctx, comodel, "", 500, filterField, filterID)
	if err != nil {
		rows = nil
	}

	phoneCodeAttr := ""
	if comodel == "core.country" && id > 0 {
		if rec, err := orm.SearchOne(ctx, "core.country", map[string]interface{}{"id": id}); err == nil {
			phoneCodeAttr = strings.TrimSpace(orm.AsString(rec["phone_code"]))
		}
	}

	sb.WriteString(`<div class="sum-field-widget sum-m2o-select"`)
	sb.WriteString(` data-sum-m2o-select`)
	sb.WriteString(` data-comodel="` + template.HTMLEscapeString(comodel) + `"`)
	sb.WriteString(` data-field="` + template.HTMLEscapeString(f.Name) + `"`)
	if parentField != "" {
		sb.WriteString(` data-parent-field="` + template.HTMLEscapeString(parentField) + `"`)
	}
	if fallbackParent != "" {
		sb.WriteString(` data-fallback-parent-field="` + template.HTMLEscapeString(fallbackParent) + `"`)
	}
	if phoneCodeAttr != "" {
		sb.WriteString(` data-phone-code="` + template.HTMLEscapeString(phoneCodeAttr) + `"`)
	}
	sb.WriteString(`>`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	if f.Name == "country_id" {
		sb.WriteString(`<div class="sum-field-phone-code" data-sum-phone-code>`)
		if phoneCodeAttr != "" {
			sb.WriteString(`+` + template.HTMLEscapeString(phoneCodeAttr))
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(fmt.Sprintf(`<select class="sum-field-input sum-field-select" name="%s" id="%s" data-sum-m2o-select-el>`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name)))
	sb.WriteString(`<option value="">—</option>`)
	for _, row := range rows {
		rid, _ := orm.CoerceInt64(row["id"])
		name := strings.TrimSpace(orm.AsString(row["name"]))
		sel := ""
		if int(rid) == id {
			sel = " selected"
		}
		extra := ""
		if comodel == "core.country" {
			if pc := strings.TrimSpace(orm.AsString(row["phone_code"])); pc != "" {
				extra = ` data-phone-code="` + template.HTMLEscapeString(pc) + `"`
			}
		}
		sb.WriteString(fmt.Sprintf(`<option value="%d"%s%s>%s</option>`, rid, sel, extra, template.HTMLEscapeString(name)))
	}
	sb.WriteString(`</select></div>`)
	_ = resModel
}

func renderOne2ManyField(ctx context.Context, sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, parentModel, comodel string, vr *ViewRecordData) {
	sb.WriteString(`<div class="sum-field-widget sum-field-widget--full sum-o2m">`)
	sb.WriteString(`<div class="sum-o2m-title">` + template.HTMLEscapeString(label) + `</div>`)
	parentID := 0
	if vr != nil {
		parentID = vr.RecordID
	}
	if parentID <= 0 {
		if id, ok := orm.CoerceInt64(record["id"]); ok {
			parentID = int(id)
		}
	}
	if parentID <= 0 || comodel == "" || parentModel == "" {
		sb.WriteString(`<p class="sum-security-muted">Save the record to manage related lines.</p></div>`)
		return
	}
	inv := orm.ResolveInverseOne2ManyField(parentModel, comodel)
	if inv == "" {
		sb.WriteString(`<p class="sum-security-muted">No inverse link for ` + template.HTMLEscapeString(comodel) + `.</p></div>`)
		return
	}
	rows, err := orm.Search(ctx, comodel, [][]interface{}{{inv, "=", parentID}})
	if err != nil {
		sb.WriteString(`<p class="sum-security-muted">` + template.HTMLEscapeString(err.Error()) + `</p></div>`)
		return
	}
	cols := o2mDisplayColumns(comodel)
	sb.WriteString(`<div class="sum-o2m-table-wrap"><table class="sum-o2m-table"><thead><tr>`)
	for _, c := range cols {
		sb.WriteString(`<th>` + template.HTMLEscapeString(c.Label) + `</th>`)
	}
	if !ro {
		sb.WriteString(`<th></th>`)
	}
	sb.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		rid, _ := orm.CoerceInt64(row["id"])
		sb.WriteString(`<tr>`)
		for _, c := range cols {
			cell := orm.AsString(row[c.Name])
			if fd := fieldDef(comodel, c.Name); fd != nil && fd.Type == orm.Many2One {
				if id, ok := orm.CoerceInt64(row[c.Name]); ok && id > 0 {
					cell = orm.DisplayNameForID(ctx, fd.Relation, int(id))
				}
			}
			sb.WriteString(`<td>` + template.HTMLEscapeString(cell) + `</td>`)
		}
		if !ro {
			sb.WriteString(`<td class="sum-o2m-muted">#` + template.HTMLEscapeString(fmt.Sprintf("%d", int(rid))) + `</td>`)
		}
		sb.WriteString(`</tr>`)
	}
	if len(rows) == 0 {
		sb.WriteString(`<tr><td colspan="8" class="sum-o2m-empty">No lines yet.</td></tr>`)
	}
	sb.WriteString(`</tbody></table></div>`)
	if !ro {
		sb.WriteString(`<p class="sum-security-muted">Add lines from the related list action or create with the inverse field set.</p>`)
	}
	sb.WriteString(`</div>`)
}

type o2mCol struct{ Name, Label string }

func o2mDisplayColumns(comodel string) []o2mCol {
	m, ok := orm.Registry[comodel]
	if !ok || m == nil {
		return nil
	}
	var cols []o2mCol
	for _, f := range m.Fields() {
		if f.Type == orm.One2Many || f.Type == orm.Many2Many {
			continue
		}
		if f.Name == "id" {
			continue
		}
		lab := f.String
		if lab == "" {
			lab = f.Name
		}
		cols = append(cols, o2mCol{Name: f.Name, Label: lab})
		if len(cols) >= 5 {
			break
		}
	}
	return cols
}
