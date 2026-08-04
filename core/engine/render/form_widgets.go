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
	if fd := fieldDef(resModel, f.Name); fd != nil {
		switch fd.Type {
		case orm.Many2One:
			renderMany2OneField(ctx, sb, f, label, record, ro, fd.Relation)
			return
		case orm.One2Many:
			renderOne2ManyField(ctx, sb, f, label, record, ro, resModel, fd.Relation, vr)
			return
		case orm.Text:
			renderTextareaField(sb, f, label, record, ro)
			return
		}
	}

	raw, hasRaw := rawField(record, f.Name)
	isBoolish := f.Widget == "boolean" || strings.HasSuffix(f.Name, "_active")
	if isBoolish {
		checked := ""
		if hasRaw && isTruthyDB(raw) {
			checked = ` checked`
		}
		dis := ""
		if ro {
			dis = ` disabled`
		}
		if ro {
			val := "No"
			if hasRaw && isTruthyDB(raw) {
				val = "Yes"
			}
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value sum-read-value--bool">` + template.HTMLEscapeString(val) + `</div>`)
			sb.WriteString(`</div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-field-checkbox" id="%s" name="%s" type="checkbox"%s%s />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), checked, dis))
		sb.WriteString(`</div>`)
		return
	}

	if f.Widget == "image" {
		src := recStr(record, f.Name)
		if ro {
			sb.WriteString(`<div class="sum-read-field sum-read-field--row">`)
			sb.WriteString(`<div class="sum-read-label">` + template.HTMLEscapeString(label) + `</div>`)
			sb.WriteString(`<div class="sum-read-value">`)
			if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
				sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-read-image" />`, template.HTMLEscapeString(src)))
			} else {
				sb.WriteString(`<span class="sum-read-empty">—</span>`)
			}
			sb.WriteString(`</div></div>`)
			return
		}
		sb.WriteString(`<div class="sum-field-widget">`)
		if src != "" && (strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "data:")) {
			sb.WriteString(`<div class="sum-image-thumb">`)
			sb.WriteString(fmt.Sprintf(`<img src="%s" alt="" class="sum-image-thumb-img" />`, template.HTMLEscapeString(src)))
			sb.WriteString(`</div>`)
		} else {
			sb.WriteString(`<div class="sum-image-placeholder">`)
			sb.WriteString(`<svg class="sum-image-placeholder-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z"></path></svg>`)
			sb.WriteString(`</div>`)
		}
		sb.WriteString(`</div>`)
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
