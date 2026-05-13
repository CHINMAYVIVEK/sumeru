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

func TestParseViewFromArch_viewRootTreeOpenFalse(t *testing.T) {
	arch := `<view type="tree" open="false"><field name="x" string="X"/></view>`
	v, err := ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "tree" || len(v.Field) != 1 {
		t.Fatalf("unexpected view: %#v", v)
	}
	if !v.TreeNoRowOpen {
		t.Fatalf("expected TreeNoRowOpen for view root with open=false")
	}
}

func TestParseViewFromArch_treeOpenFalse(t *testing.T) {
	arch := `<tree open="false"><field name="a"/></tree>`
	v, err := ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if !v.TreeNoRowOpen {
		t.Fatalf("expected TreeNoRowOpen true for open=false, got %#v", v)
	}
}

func TestParseViewFromArch_treeOpenDefault(t *testing.T) {
	arch := `<tree><field name="a"/></tree>`
	v, err := ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.TreeNoRowOpen {
		t.Fatalf("expected TreeNoRowOpen false by default, got %#v", v)
	}
}
