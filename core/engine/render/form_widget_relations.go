package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

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
		sb.WriteString(`<div class="sum-field-widget">`)
		sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" type="text" value="%s" readonly />`,
			template.HTMLEscapeString(f.Name), template.HTMLEscapeString(display)))
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
		renderReadonlyFieldValue(sb, f, label, display)
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
	found := false
	for _, row := range rows {
		rid, _ := orm.CoerceInt64(row["id"])
		name := strings.TrimSpace(orm.AsString(row["name"]))
		sel := ""
		if int(rid) == id {
			sel = " selected"
			found = true
		}
		extra := ""
		if comodel == "core.country" {
			if pc := strings.TrimSpace(orm.AsString(row["phone_code"])); pc != "" {
				extra = ` data-phone-code="` + template.HTMLEscapeString(pc) + `"`
			}
		}
		sb.WriteString(fmt.Sprintf(`<option value="%d"%s%s>%s</option>`, rid, sel, extra, template.HTMLEscapeString(name)))
	}
	if id > 0 && !found && display != "" {
		sb.WriteString(fmt.Sprintf(`<option value="%d" selected>%s</option>`, id, template.HTMLEscapeString(display)))
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
