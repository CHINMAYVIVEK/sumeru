package parser

import "testing"

func TestParseViewFromArch_formRoot(t *testing.T) {
	arch := `<form><header><button name="save" string="Save" type="object"/></header><sheet><group><field name="x" string="X"/></group></sheet></form>`
	v, err := ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "form" || v.Header == nil || v.Sheet == nil {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func TestParseViewFromArch_treeRoot(t *testing.T) {
	arch := `<tree><field name="a"/><field name="b" string="B"/></tree>`
	v, err := ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "tree" || len(v.Field) != 2 {
		t.Fatalf("unexpected view: %#v", v)
	}
}
