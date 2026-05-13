package orm

import "testing"

func TestColumnTypeSQL(t *testing.T) {
	_, ok := ColumnTypeSQL(FieldDefinition{Name: "x", Type: FieldType("nope")})
	if ok {
		t.Fatal("unknown type should return ok=false")
	}
	s, ok := ColumnTypeSQL(FieldDefinition{Name: "m", Type: Many2One})
	if !ok || s != "BIGINT" {
		t.Fatalf("many2one: got %q %v", s, ok)
	}
}

func TestBuildAddColumnDefinitionBooleanDefault(t *testing.T) {
	got := buildAddColumnDefinition(FieldDefinition{Name: "b", Type: Boolean, DefaultVal: true}, "BOOLEAN")
	if got != "BOOLEAN NOT NULL DEFAULT TRUE" {
		t.Fatalf("got %q", got)
	}
}
