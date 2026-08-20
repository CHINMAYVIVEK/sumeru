package orm

import (
	"strings"
	"testing"
)

func TestBuildSearchWhereClauseDateNotFalseIsNotNull(t *testing.T) {
	registerStubModel(t, "test.date.domain", []FieldDefinition{
		{Name: "date_deadline", Type: Date},
		{Name: "name", Type: Char},
	})

	where, args, err := buildSearchWhereClause("test.date.domain", [][]interface{}{
		{"date_deadline", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if where == "" || !strings.Contains(where, "IS NOT NULL") {
		t.Fatalf("expected IS NOT NULL, got %q", where)
	}
}

func TestBuildSearchWhereClauseDateEqualsFalseIsNull(t *testing.T) {
	registerStubModel(t, "test.date.domain.eq", []FieldDefinition{
		{Name: "date_deadline", Type: Date},
	})

	where, args, err := buildSearchWhereClause("test.date.domain.eq", [][]interface{}{
		{"date_deadline", "=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 0 {
		t.Fatalf("expected no args, got %v", args)
	}
	if where == "" || !strings.Contains(where, "IS NULL") {
		t.Fatalf("expected IS NULL, got %q", where)
	}
}

func TestBuildSearchWhereClauseNonDateFalseUsesDistinctFrom(t *testing.T) {
	registerStubModel(t, "test.bool.domain", []FieldDefinition{
		{Name: "active", Type: Boolean},
	})

	where, args, err := buildSearchWhereClause("test.bool.domain", [][]interface{}{
		{"active", "!=", false},
	})
	if err != nil {
		t.Fatalf("buildSearchWhereClause: %v", err)
	}
	if len(args) != 1 || args[0] != false {
		t.Fatalf("expected one false arg, got args=%v", args)
	}
	if strings.Contains(where, "IS NOT NULL") {
		t.Fatalf("boolean field should not use IS NOT NULL, got %q", where)
	}
	if !strings.Contains(where, "IS DISTINCT FROM") {
		t.Fatalf("expected IS DISTINCT FROM, got %q", where)
	}
}
