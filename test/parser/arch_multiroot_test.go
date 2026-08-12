package parser_test

import (
	"os"
	"testing"

	"sumeru/core/engine/parser"
)

func TestParseViewFromArch_formRoot(t *testing.T) {
	arch := `<form><header><button name="save" string="Save" type="object"/></header><sheet><group><field name="x" string="X"/></group></sheet></form>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "form" || v.Header == nil || v.Sheet == nil {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func TestParseViewFromArch_nestedFormUnderView(t *testing.T) {
	arch := `<view id="v1" model="demo.model" type="form"><form><sheet><field name="x" string="X"/></sheet></form></view>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Sheet == nil || len(v.Sheet.Field) != 1 || v.Sheet.Field[0].Name != "x" {
		t.Fatalf("expected nested form sheet/field promoted, got %#v", v)
	}
	if v.Form != nil {
		t.Fatalf("expected Form cleared after promote, got %#v", v.Form)
	}
}

func TestParseViewList_nestedFormUnderView(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/form.xml"
	content := `<?xml version="1.0" encoding="utf-8"?>
<sumeru>
  <data>
    <view id="view_demo_form" model="demo.model" type="form">
      <form>
        <sheet>
          <field name="x" string="X"/>
        </sheet>
      </form>
    </view>
  </data>
</sumeru>
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	vl, err := parser.ParseViewList(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(vl.Views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(vl.Views))
	}
	v := vl.Views[0]
	if v.Sheet == nil || len(v.Sheet.Field) != 1 || v.Sheet.Field[0].Name != "x" {
		t.Fatalf("expected nested form promoted in ParseViewList, got %#v", v)
	}
	if v.Form != nil {
		t.Fatalf("expected Form cleared after promote, got %#v", v.Form)
	}
}

func TestParseViewFromArch_treeRoot(t *testing.T) {
	arch := `<tree><field name="a"/><field name="b" string="B"/></tree>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != "tree" || len(v.Field) != 2 {
		t.Fatalf("unexpected view: %#v", v)
	}
}

func TestParseViewFromArch_viewRootTreeOpenFalse(t *testing.T) {
	arch := `<view type="tree" open="false"><field name="x" string="X"/></view>`
	v, err := parser.ParseViewFromArch(arch)
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
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if !v.TreeNoRowOpen {
		t.Fatalf("expected TreeNoRowOpen true for open=false, got %#v", v)
	}
}

func TestParseViewFromArch_treeOpenDefault(t *testing.T) {
	arch := `<tree><field name="a"/></tree>`
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		t.Fatal(err)
	}
	if v.TreeNoRowOpen {
		t.Fatalf("expected TreeNoRowOpen false by default, got %#v", v)
	}
}
