package swcmeta

import (
	"strings"
	"testing"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func TestSerializeSheetNestedGroupsAndDivs(t *testing.T) {
	sheet := &parser.Sheet{
		Div: []parser.Div{
			{
				Class: "sum_title",
				H1: []parser.H1{
					{Field: []parser.Field{{Name: "name", Label: "Name"}}},
				},
			},
		},
		Group: []parser.Group{
			{
				Group: []parser.Group{
					{
						Title: "Contact",
						Field: []parser.Field{{Name: "email", Label: "Email"}},
					},
					{
						Title: "Address",
						Field: []parser.Field{{Name: "street", Label: "Street"}},
					},
				},
			},
		},
		Notebook: []parser.Notebook{
			{
				Page: []parser.Page{
					{
						Title: "Notes",
						Field: []parser.Field{{Name: "comment", Label: "Notes"}},
					},
				},
			},
		},
	}

	got := serializeSheet("core.partner", sheet)
	if got == nil {
		t.Fatal("expected sheet")
	}
	if len(got.Divs) != 1 || len(got.Divs[0].H1Fields) != 1 || got.Divs[0].H1Fields[0].Name != "name" {
		t.Fatalf("unexpected divs: %+v", got.Divs)
	}
	if len(got.Groups) != 1 || len(got.Groups[0].Groups) != 2 {
		t.Fatalf("expected nested groups: %+v", got.Groups)
	}
	if got.Groups[0].Groups[0].String != "Contact" || got.Groups[0].Groups[0].Fields[0].Name != "email" {
		t.Fatalf("unexpected contact group: %+v", got.Groups[0].Groups[0])
	}
	if len(got.Notebook) != 1 || len(got.Notebook[0].Pages) != 1 {
		t.Fatalf("unexpected notebook: %+v", got.Notebook)
	}
	if got.Notebook[0].Pages[0].Fields[0].Name != "comment" {
		t.Fatalf("unexpected page fields: %+v", got.Notebook[0].Pages[0].Fields)
	}
}

func TestSerializeGroupRecursive(t *testing.T) {
	g := serializeGroup("core.partner", parser.Group{
		Title: "Outer",
		Group: []parser.Group{
			{Title: "Inner", Field: []parser.Field{{Name: "x"}}},
		},
	})
	if g.String != "Outer" || len(g.Groups) != 1 || g.Groups[0].Fields[0].Name != "x" {
		t.Fatalf("unexpected group: %+v", g)
	}
	if strings.TrimSpace(g.Groups[0].String) != "Inner" {
		t.Fatalf("expected inner title")
	}
}

func TestSerializeDivContactRow(t *testing.T) {
	div := serializeDiv("core.user", parser.Div{
		Class: "sum_title",
		H1: []parser.H1{
			{Field: []parser.Field{{Name: "name", Placeholder: "Name"}}},
		},
		Div: []parser.Div{
			{
				Class: "sum-title-contact-row",
				Field: []parser.Field{
					{Name: "email", Label: "Email", Widget: "email"},
					{Name: "phone", Label: "Phone"},
				},
			},
		},
	})
	if len(div.H1Fields) != 1 || div.H1Fields[0].Name != "name" {
		t.Fatalf("unexpected h1: %+v", div.H1Fields)
	}
	if len(div.Divs) != 1 || len(div.Divs[0].Fields) != 2 {
		t.Fatalf("unexpected contact row: %+v", div.Divs)
	}
}

func TestSerializeSheetSeparatorsAndLabels(t *testing.T) {
	sheet := &parser.Sheet{
		Separator: []parser.Separator{{String: "Section"}},
		Label:     []parser.Label{{For: "email", String: "Email hint"}},
	}
	got := serializeSheet("my.module", sheet)
	if len(got.Separators) != 1 || got.Separators[0].String != "Section" {
		t.Fatalf("unexpected separators: %+v", got.Separators)
	}
	if len(got.Labels) != 1 || got.Labels[0].For != "email" {
		t.Fatalf("unexpected labels: %+v", got.Labels)
	}
}

func TestFormMetaHasImageField(t *testing.T) {
	const model = "test.formmeta.partner"
	orm.Registry[model] = stubImageModel{}
	t.Cleanup(func() { delete(orm.Registry, model) })

	meta := formMetaForModel(model)
	if meta == nil || !meta.HasImageField {
		t.Fatal("expected model to have image field")
	}
}

type stubImageModel struct{ orm.BaseModel }

func (stubImageModel) ModelName() string { return "test.formmeta.partner" }

func (stubImageModel) Fields() []orm.FieldDefinition {
	return []orm.FieldDefinition{{Name: "image", Type: orm.Text}}
}
