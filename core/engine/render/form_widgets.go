package render

import (
	"context"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// form_widgets.go dispatches arch fields to scalar and relation widget renderers.

func renderField(ctx context.Context, sb *strings.Builder, f parser.Field, record map[string]interface{}, ro bool, vr *ViewRecordData) {
	if gs := strings.TrimSpace(f.Groups); gs != "" {
		if !orm.UserHasAnyAccessGroup(ctx, orm.SecurityUID(ctx), gs) {
			return
		}
	}
	label := FieldDisplayLabel(f)

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
	if f.Widget == "priority" {
		renderPriorityField(sb, f, label, record, ro)
		return
	}
	if f.Widget == "many2many_tags" {
		txt := recStr(record, f.Name)
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
