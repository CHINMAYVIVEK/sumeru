package parser_test

import (
	"encoding/xml"
	"testing"

	"sumeru/core/engine/parser"
)

func TestModuleXML_dataWrapperSumeru(t *testing.T) {
	in := []byte(`<sumeru><data><record id="x" model="y"><field name="a">b</field></record></data></sumeru>`)
	root, err := parser.PeekModuleXMLRootName(in)
	if err != nil || root != "sumeru" {
		t.Fatalf("root %q err %v", root, err)
	}
	if err := parser.ValidateModuleRoot(xml.Name{Local: root}); err != nil {
		t.Fatal(err)
	}
	var vl parser.ViewList
	if err := xml.Unmarshal(in, &vl); err != nil {
		t.Fatal(err)
	}
	vl.MergeViewListData()
	if len(vl.Records) != 1 || vl.Records[0].ID != "x" {
		t.Fatalf("records: %+v", vl.Records)
	}
}

func TestPeekModuleXMLRootName_sumeru(t *testing.T) {
	root, err := parser.PeekModuleXMLRootName([]byte("\n  <sumeru>\n</sumeru>"))
	if err != nil || root != "sumeru" {
		t.Fatalf("got %q err %v", root, err)
	}
}
