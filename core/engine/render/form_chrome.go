package render

import (
	"context"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func renderHeader(ctx context.Context, sb *strings.Builder, h *parser.Header, record map[string]interface{}, vr *ViewRecordData) {
	sb.WriteString(`<div class="sum-form-statusbar">`)

	if len(h.Button) > 0 {
		sb.WriteString(`<div class="sum-statusbar-buttons">`)
		model := ""
		id := 0
		next := ""
		if vr != nil {
			model = strings.TrimSpace(vr.ResModel)
			id = vr.RecordID
			if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
				next = "/web?" + q
			}
		}
		csrf := ""
		if vr != nil {
			csrf = vr.CSRFToken
		}
		for _, b := range h.Button {
			writeObjectActionButton(sb, b, model, id, next, csrf, false)
		}
		sb.WriteString(`</div>`)
	}

	sb.WriteString(`<div class="sum-statusbar-status">`)
	writeStatusbarChips(ctx, sb, h, record, vr)
	sb.WriteString(`</div>`)

	sb.WriteString(`</div>`)
}

func writeStatusbarChips(ctx context.Context, sb *strings.Builder, h *parser.Header, record map[string]interface{}, vr *ViewRecordData) {
	if h == nil {
		return
	}
	for _, hf := range h.Field {
		if hf.Widget == "statusbar" && isStatusbarClickable(hf) {
			renderClickableStatusbar(ctx, sb, hf, record, vr)
			continue
		}
		val := recStr(record, hf.Name)
		if val == "" {
			continue
		}
		sb.WriteString(`<span class="sum-statusbar-chip">`)
		sb.WriteString(template.HTMLEscapeString(val))
		sb.WriteString(`</span>`)
	}
}

func isStatusbarClickable(f parser.Field) bool {
	opt := strings.ToLower(strings.TrimSpace(f.Options))
	return strings.Contains(opt, "clickable")
}

func renderClickableStatusbar(ctx context.Context, sb *strings.Builder, f parser.Field, record map[string]interface{}, vr *ViewRecordData) {
	fd := fieldDef(vrResModel(vr), f.Name)
	if fd == nil || fd.Type != orm.Many2One || fd.Relation == "" {
		val := recStr(record, f.Name)
		if val != "" {
			sb.WriteString(`<span class="sum-statusbar-chip">` + template.HTMLEscapeString(val) + `</span>`)
		}
		return
	}
	currentID := int64(0)
	if raw, ok := rawField(record, f.Name); ok {
		currentID, _ = orm.CoerceInt64(raw)
	}
	rows, err := orm.Search(ctx, fd.Relation, nil)
	if err != nil || len(rows) == 0 {
		return
	}
	sort.SliceStable(rows, func(i, j int) bool {
		si, _ := orm.CoerceInt64(rows[i]["sequence"])
		sj, _ := orm.CoerceInt64(rows[j]["sequence"])
		if si != sj {
			return si < sj
		}
		return orm.AsString(rows[i]["name"]) < orm.AsString(rows[j]["name"])
	})
	model := vrResModel(vr)
	recordID := 0
	csrf := ""
	if vr != nil {
		recordID = vr.RecordID
		csrf = vr.CSRFToken
	}
	sb.WriteString(`<div class="sum-statusbar-stages" data-sum-statusbar`)
	sb.WriteString(` data-model="` + template.HTMLEscapeString(model) + `"`)
	sb.WriteString(` data-record-id="` + template.HTMLEscapeString(fmt.Sprintf("%d", recordID)) + `"`)
	sb.WriteString(` data-field="` + template.HTMLEscapeString(f.Name) + `"`)
	sb.WriteString(` data-csrf="` + template.HTMLEscapeString(csrf) + `"`)
	sb.WriteString(`>`)
	for _, row := range rows {
		rid, ok := orm.CoerceInt64(row["id"])
		if !ok || rid <= 0 {
			continue
		}
		cls := "sum-statusbar-stage"
		if rid == currentID {
			cls += " sum-statusbar-stage--current"
		}
		sb.WriteString(`<button type="button" class="` + cls + `" data-stage-id="` + template.HTMLEscapeString(fmt.Sprintf("%d", rid)) + `">`)
		sb.WriteString(template.HTMLEscapeString(orm.AsString(row["name"])))
		sb.WriteString(`</button>`)
	}
	sb.WriteString(`</div>`)
}

func vrResModel(vr *ViewRecordData) string {
	if vr == nil {
		return ""
	}
	return strings.TrimSpace(vr.ResModel)
}

func renderButtonBox(sb *strings.Builder, d parser.Div) {
	// Empty button-box: omit spacer chrome (no dead vertical gap).
	if len(d.Field) == 0 && len(d.H1) == 0 {
		return
	}
	sb.WriteString(`<div class="sum-form-toolbar-spacer" aria-hidden="true"></div>`)
}

func renderTitleAvatar(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool, resModel string) {
	_ = ctx
	_ = d
	img := strings.TrimSpace(recStr(record, "image"))
	cropRaw := strings.TrimSpace(recStr(record, "image_crop"))
	crop, hasCrop := ParseImageCrop(cropRaw)
	cropStyle := AvatarCropStyle(crop, hasCrop && SafeImageSrc(img))
	name := strings.TrimSpace(recStr(record, "name"))
	initials := UserInitialsFromName(name)
	canUpload := fieldDef(resModel, "image") != nil
	hasCropField := fieldDef(resModel, "image_crop") != nil

	hasImg := SafeImageSrc(img)

	imgClass := "sum-form-avatar-img"
	if hasImg {
		imgClass += " sum-form-avatar-img--visible"
		if hasCrop {
			imgClass += " sum-form-avatar-img--cropped"
		}
	}

	sb.WriteString(`<div class="sum-form-avatar sum-form-avatar--compact" data-sum-avatar>`)
	sb.WriteString(`<div class="sum-form-avatar-box sum-form-avatar-box--circle">`)
	if hasImg {
		sb.WriteString(fmt.Sprintf(`<img class="%s" data-sum-avatar-preview src="%s" alt=""%s />`,
			imgClass, template.HTMLEscapeString(img), cropStyle))
		sb.WriteString(`<span class="sum-form-avatar-initials" data-sum-avatar-initials hidden>` + template.HTMLEscapeString(initials) + `</span>`)
	} else {
		sb.WriteString(`<span class="sum-form-avatar-initials" data-sum-avatar-initials>` + template.HTMLEscapeString(initials) + `</span>`)
		sb.WriteString(`<img class="sum-form-avatar-img" data-sum-avatar-preview alt="" hidden />`)
	}
	sb.WriteString(`</div>`)
	if !ro && canUpload {
		sb.WriteString(fmt.Sprintf(`<input type="hidden" name="image" value="%s" data-sum-avatar-value />`, template.HTMLEscapeString(img)))
		if hasCropField {
			sb.WriteString(fmt.Sprintf(`<input type="hidden" name="image_crop" value="%s" data-sum-avatar-crop />`, template.HTMLEscapeString(cropRaw)))
		}
		sb.WriteString(`<div class="sum-form-avatar-actions">`)
		sb.WriteString(`<label class="sum-form-avatar-upload">`)
		sb.WriteString(`<input type="file" accept="image/*" data-sum-avatar-file />`)
		sb.WriteString(`<span>Change</span>`)
		sb.WriteString(`</label>`)
		if hasImg {
			sb.WriteString(`<button type="button" class="sum-form-avatar-adjust" data-sum-avatar-adjust>Adjust</button>`)
		}
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div>`)
}

func renderTitleBody(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool) {
	_ = ctx
	sb.WriteString(`<div class="sum-form-title-body sum-form-title-body--main">`)
	roAttr := ""
	if ro {
		roAttr = ` readonly`
	}
	for _, h1 := range d.H1 {
		for _, f := range h1.Field {
			ph := template.HTMLEscapeString(f.Placeholder)
			if ph == "" {
				ph = template.HTMLEscapeString(f.Label)
			}
			if ph == "" {
				ph = "Name"
			}
			v := recStr(record, f.Name)
			sb.WriteString(fmt.Sprintf(`<input class="sum-form-hero-input sum-form-hero-input--bold" placeholder="%s" name="%s" value="%s"%s />`,
				ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v), roAttr))
		}
	}
	sb.WriteString(`<div class="sum-form-contact-row">`)
	for _, f := range d.Field {
		v := recStr(record, f.Name)
		ph := template.HTMLEscapeString(f.Label)
		if ph == "" {
			ph = template.HTMLEscapeString(f.Name)
		}
		inputType := "text"
		if f.Widget == "email" || strings.Contains(strings.ToLower(f.Name), "email") {
			inputType = "email"
		} else if f.Widget == "phone" || f.Widget == "tel" || strings.Contains(strings.ToLower(f.Name), "phone") || strings.Contains(strings.ToLower(f.Name), "mobile") {
			inputType = "tel"
		}
		sb.WriteString(`<div class="sum-form-contact-item">`)
		sb.WriteString(fmt.Sprintf(`<input class="sum-form-inline-input" type="%s" placeholder="%s" name="%s" value="%s"%s />`,
			inputType, ph, template.HTMLEscapeString(f.Name), template.HTMLEscapeString(v), roAttr))
		sb.WriteString(`</div>`)
	}
	sb.WriteString(`</div></div>`)
}

func renderTitle(ctx context.Context, sb *strings.Builder, d parser.Div, record map[string]interface{}, ro bool, resModel string) {
	sb.WriteString(`<div class="sum-form-title-row sum-form-title-row--sheet">`)
	if fieldDef(resModel, "image") != nil {
		renderTitleAvatar(ctx, sb, d, record, ro, resModel)
	}
	renderTitleBody(ctx, sb, d, record, ro)
	sb.WriteString(`</div>`)
}

func renderFormFooter(sb *strings.Builder, ft *parser.Footer, vr *ViewRecordData) {
	if ft == nil || len(ft.Button) == 0 {
		return
	}
	model := ""
	id := 0
	next := ""
	if vr != nil {
		model = strings.TrimSpace(vr.ResModel)
		id = vr.RecordID
		if q := strings.TrimSpace(vr.FormBaseQuery); q != "" {
			next = "/web?" + q
		}
	}
	csrf := ""
	if vr != nil {
		csrf = vr.CSRFToken
	}
	sb.WriteString(`<div class="sum-form-footer" role="group" aria-label="Form actions">`)
	for _, b := range ft.Button {
		writeObjectActionButton(sb, b, model, id, next, csrf, true)
	}
	sb.WriteString(`</div>`)
}
