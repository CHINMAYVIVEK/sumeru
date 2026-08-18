package swcmeta

import (
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// SerializeView converts a parsed view arch into SWC JSON metadata.
func SerializeView(view *parser.View) ViewArch {
	if view == nil {
		return ViewArch{}
	}
	model := strings.TrimSpace(view.Model)
	arch := ViewArch{
		Type:   strings.TrimSpace(view.Type),
		Model:  model,
		Title:  strings.TrimSpace(view.Title),
		Fields: enrichFields(model, serializeFields(view.Field)),
	}
	if view.Header != nil {
		arch.Header = &ArchHeader{
			Buttons: serializeButtons(view.Header.Button),
			Fields:  enrichFields(model, serializeFields(view.Header.Field)),
		}
	}
	if view.Footer != nil && len(view.Footer.Button) > 0 {
		arch.Footer = &ArchFooter{
			Buttons: serializeButtons(view.Footer.Button),
		}
	}
	if view.Sheet != nil {
		arch.Sheet = serializeSheet(model, view.Sheet)
	}
	arch.FormMeta = formMetaForModel(model)
	if strings.EqualFold(view.Type, "kanban") {
		arch.Kanban = &KanbanMeta{
			GroupField: view.KanbanGroupField(),
			Draggable:  view.KanbanDraggable(),
		}
	}
	if rep := view.Report; rep != nil {
		arch.Report = &ReportMeta{
			Download:  rep.Download != "" && rep.Download != "0" && rep.Download != "false",
			Upload:    rep.Upload != "" && rep.Upload != "0" && rep.Upload != "false",
			PDFSizes:  rep.PDFSizes,
			BulkModes: rep.Modes,
		}
	}
	return arch
}

func formMetaForModel(model string) *FormMeta {
	inst, ok := orm.Registry[model]
	if !ok {
		return &FormMeta{}
	}
	meta := &FormMeta{}
	for _, f := range inst.Fields() {
		if f.Name == "image" {
			meta.HasImageField = true
			break
		}
	}
	return meta
}

func enrichFields(model string, fields []ArchField) []ArchField {
	if model == "" || len(fields) == 0 {
		return fields
	}
	out := make([]ArchField, len(fields))
	for i, f := range fields {
		out[i] = enrichField(model, f)
	}
	return out
}

func enrichField(model string, f ArchField) ArchField {
	inst, ok := orm.Registry[model]
	if !ok {
		return f
	}
	for _, fd := range inst.Fields() {
		if fd.Name != f.Name {
			continue
		}
		if f.Type == "" {
			f.Type = string(fd.Type)
		}
		if f.String == "" {
			f.String = fd.String
		}
		if f.Relation == "" {
			f.Relation = fd.Relation
		}
		if len(f.Selection) == 0 && len(fd.Selection) > 0 {
			f.Selection = fd.Selection
		}
		if fd.Required {
			f.Required = true
		}
		if f.Options == nil && fd.Relation != "" {
			f.Options = map[string]string{"relation": fd.Relation}
		}
		break
	}
	return f
}

func serializeSheet(model string, s *parser.Sheet) *ArchSheet {
	out := &ArchSheet{
		Fields:     enrichFields(model, serializeFields(s.Field)),
		Groups:     []ArchGroup{},
		Divs:       serializeDivs(model, s),
		Separators: serializeSeparators(s.Separator),
		Labels:     serializeLabels(s.Label),
	}
	for _, g := range s.Group {
		out.Groups = append(out.Groups, serializeGroup(model, g))
	}
	for _, nb := range s.Notebook {
		pages := make([]ArchPage, 0, len(nb.Page))
		for _, p := range nb.Page {
			pg := ArchPage{
				Title:      strings.TrimSpace(p.Title),
				Fields:     enrichFields(model, serializeFields(p.Field)),
				Groups:     []ArchGroup{},
				Separators: serializeSeparators(p.Separator),
				Labels:     serializeLabels(p.Label),
			}
			for _, g := range p.Group {
				pg.Groups = append(pg.Groups, serializeGroup(model, g))
			}
			pages = append(pages, pg)
		}
		out.Notebook = append(out.Notebook, ArchNotebook{Pages: pages})
	}
	return out
}

func serializeGroup(model string, g parser.Group) ArchGroup {
	out := ArchGroup{
		String:     strings.TrimSpace(g.Title),
		Fields:     enrichFields(model, serializeFields(g.Field)),
		Groups:     []ArchGroup{},
		Separators: serializeSeparators(g.Separator),
		Labels:     serializeLabels(g.Label),
	}
	for _, nested := range g.Group {
		out.Groups = append(out.Groups, serializeGroup(model, nested))
	}
	return out
}

func serializeDivs(model string, s *parser.Sheet) []ArchDiv {
	if s == nil || len(s.Div) == 0 {
		return nil
	}
	out := make([]ArchDiv, 0, len(s.Div))
	for _, d := range s.Div {
		div := serializeDiv(model, d)
		if divHasContent(div) {
			out = append(out, div)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func serializeDiv(model string, d parser.Div) ArchDiv {
	out := ArchDiv{
		Class:  strings.TrimSpace(d.Class),
		Fields: enrichFields(model, serializeFields(d.Field)),
	}
	for _, h1 := range d.H1 {
		out.H1Fields = append(out.H1Fields, enrichFields(model, serializeFields(h1.Field))...)
	}
	for _, nested := range d.Div {
		child := serializeDiv(model, nested)
		if divHasContent(child) {
			out.Divs = append(out.Divs, child)
		}
	}
	return out
}

func divHasContent(d ArchDiv) bool {
	return len(d.Fields) > 0 || len(d.H1Fields) > 0 || len(d.Divs) > 0
}

func serializeButtons(buttons []parser.Button) []ArchButton {
	out := make([]ArchButton, 0, len(buttons))
	for _, b := range buttons {
		out = append(out, ArchButton{
			Name:   strings.TrimSpace(b.Name),
			String: strings.TrimSpace(b.String),
			Type:   strings.TrimSpace(b.Type),
			Class:  strings.TrimSpace(b.Class),
		})
	}
	return out
}

func serializeFields(fields []parser.Field) []ArchField {
	out := make([]ArchField, 0, len(fields))
	for _, f := range fields {
		out = append(out, ArchField{
			Name:        strings.TrimSpace(f.Name),
			String:      strings.TrimSpace(f.Label),
			Widget:      strings.TrimSpace(f.Widget),
			Placeholder: strings.TrimSpace(f.Placeholder),
			PivotType:   strings.TrimSpace(f.PivotType),
			Options:     parseFieldOptions(f.Options),
		})
	}
	return out
}

func parseFieldOptions(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	out := map[string]string{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if kv := strings.SplitN(part, ":", 2); len(kv) == 2 {
			out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func serializeSeparators(items []parser.Separator) []ArchSeparator {
	if len(items) == 0 {
		return nil
	}
	out := make([]ArchSeparator, 0, len(items))
	for _, s := range items {
		out = append(out, ArchSeparator{String: strings.TrimSpace(s.String)})
	}
	return out
}

func serializeLabels(items []parser.Label) []ArchLabel {
	if len(items) == 0 {
		return nil
	}
	out := make([]ArchLabel, 0, len(items))
	for _, l := range items {
		out = append(out, ArchLabel{
			For:    strings.TrimSpace(l.For),
			String: strings.TrimSpace(l.String),
		})
	}
	return out
}
