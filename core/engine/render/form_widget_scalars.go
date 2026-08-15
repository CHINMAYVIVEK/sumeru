package render

import (
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func renderBooleanField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	raw, hasRaw := rawField(record, f.Name)
	truthy := hasRaw && isTruthyDB(raw)
	if strings.EqualFold(f.Widget, "radio") {
		renderBooleanRadio(sb, f, label, truthy, ro)
		return
	}
	renderBooleanSelect(sb, f, label, truthy, ro)
}

func renderBooleanSelect(sb *strings.Builder, f parser.Field, label string, truthy, ro bool) {
	trueLabel, falseLabel := "Yes", "No"
	if f.Name == "active" {
		trueLabel, falseLabel = "Active", "Inactive"
	}
	if ro {
		display := falseLabel
		if truthy {
			display = trueLabel
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		writeOption(sb, "true", trueLabel, truthy)
		writeOption(sb, "false", falseLabel, !truthy)
	})
}

func renderBooleanRadio(sb *strings.Builder, f parser.Field, label string, truthy, ro bool) {
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
	roAttr := ""
	if ro {
		roAttr = ` disabled`
	}
	sb.WriteString(fmt.Sprintf(`<label class="sum-field-radio"><input type="radio" name="%s" value="true"%s%s /><span>Yes</span></label>`, name, yesChecked, roAttr))
	sb.WriteString(fmt.Sprintf(`<label class="sum-field-radio"><input type="radio" name="%s" value="false"%s%s /><span>No</span></label>`, name, noChecked, roAttr))
	sb.WriteString(`</div></div>`)
}

func renderModelSelectionSelect(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, opts [][]string) {
	cur := strings.TrimSpace(recStr(record, f.Name))
	if ro {
		display := cur
		for _, o := range opts {
			if len(o) >= 2 && o[0] == cur {
				display = o[1]
				break
			}
		}
		renderReadonlyFieldValue(sb, f, label, display)
		return
	}
	renderSelectShell(sb, f, label, false, func() {
		writeOption(sb, "", "—", cur == "")
		for _, o := range opts {
			if len(o) < 2 {
				continue
			}
			writeOption(sb, o[0], o[1], o[0] == cur)
		}
	})
}

func imageSrcOK(src string) bool { return SafeImageSrc(src) }

func renderPriorityField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	cur := int64(1)
	if raw, ok := rawField(record, f.Name); ok {
		if n, ok := orm.CoerceInt64(raw); ok {
			cur = n
		}
	}
	sb.WriteString(`<div class="sum-field-widget sum-priority-field">`)
	sb.WriteString(`<span class="sum-field-label">` + template.HTMLEscapeString(label) + `</span>`)
	sb.WriteString(`<div class="sum-priority-stars" role="group" aria-label="` + template.HTMLEscapeString(label) + `">`)
	name := template.HTMLEscapeString(f.Name)
	for i := int64(0); i <= 3; i++ {
		cls := "sum-priority-star"
		if i <= cur {
			cls += " sum-priority-star--on"
		}
		if ro {
			sb.WriteString(`<span class="` + cls + `" aria-hidden="true">★</span>`)
			continue
		}
		sb.WriteString(fmt.Sprintf(`<button type="button" class="%s" data-priority="%d" data-field="%s" aria-label="Priority %d">★</button>`, cls, i, name, i))
	}
	if !ro {
		sb.WriteString(fmt.Sprintf(`<input type="hidden" name="%s" value="%d" data-sum-priority-value />`, name, cur))
	}
	sb.WriteString(`</div></div>`)
}

func renderImageField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool) {
	src := strings.TrimSpace(recStr(record, f.Name))
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
		if !ro {
			sb.WriteString(`<div class="sum-image-thumb" hidden>`)
			sb.WriteString(`<img src="" alt="" class="sum-image-thumb-img" data-sum-image-preview hidden />`)
			sb.WriteString(`</div>`)
		}
	}
	sb.WriteString(fmt.Sprintf(`<input type="hidden" id="%s" name="%s" value="%s" data-sum-image-value />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name), template.HTMLEscapeString(src)))
	if !ro {
		sb.WriteString(`<label class="sum-form-avatar-upload sum-image-upload">`)
		sb.WriteString(`<input type="file" accept="image/*" data-sum-image-file />`)
		sb.WriteString(`<span>Change</span>`)
		sb.WriteString(`</label>`)
	}
	sb.WriteString(`</div>`)
}

func renderTypedInput(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, inputType string, required bool) {
	placeholder := f.Placeholder
	val := recStr(record, f.Name)
	roAttr := ""
	if ro {
		roAttr = ` readonly`
	}
	attrs := fieldInputAttrs(label, required)
	sb.WriteString(`<div class="sum-field-widget">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<input class="sum-field-input" id="%s" name="%s" type="%s" placeholder="%s" value="%s"%s%s />`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(inputType),
		template.HTMLEscapeString(placeholder), template.HTMLEscapeString(val), roAttr, attrs))
	sb.WriteString(`</div>`)
}

func renderTextareaField(sb *strings.Builder, f parser.Field, label string, record map[string]interface{}, ro bool, required bool) {
	val := recStr(record, f.Name)
	placeholder := f.Placeholder
	roAttr := ""
	if ro {
		roAttr = ` readonly`
	}
	attrs := fieldInputAttrs(label, required)
	sb.WriteString(`<div class="sum-field-widget sum-field-widget--full">`)
	sb.WriteString(`<label class="sum-field-label" for="` + template.HTMLEscapeString(f.Name) + `">` + template.HTMLEscapeString(label) + `</label>`)
	sb.WriteString(fmt.Sprintf(`<textarea class="sum-field-input sum-field-textarea" id="%s" name="%s" rows="4" placeholder="%s"%s%s>%s</textarea>`,
		template.HTMLEscapeString(f.Name), template.HTMLEscapeString(f.Name),
		template.HTMLEscapeString(placeholder), roAttr, attrs, template.HTMLEscapeString(val)))
	sb.WriteString(`</div>`)
}
