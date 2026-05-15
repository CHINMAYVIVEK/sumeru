package orm_test

import (
	"testing"

	"sumeru/core/orm"
)

func TestColumnTypeSQL(t *testing.T) {
	_, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "x", Type: orm.FieldType("nope")})
	if ok {
		t.Fatal("unknown type should return ok=false")
	}
	s, ok := orm.ColumnTypeSQL(orm.FieldDefinition{Name: "m", Type: orm.Many2One})
	if !ok || s != "BIGINT" {
		t.Fatalf("many2one: got %q %v", s, ok)
	}
}

func TestFormatAddColumnDefinitionBooleanDefault(t *testing.T) {
	got := orm.FormatAddColumnDefinition(orm.FieldDefinition{Name: "b", Type: orm.Boolean, DefaultVal: true}, "BOOLEAN")
	if got != "BOOLEAN NOT NULL DEFAULT TRUE" {
		t.Fatalf("got %q", got)
	}
}
