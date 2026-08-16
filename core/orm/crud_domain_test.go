package orm

import (
	"strings"
	"testing"
)

type testDateDomainModel struct {
	BaseModel
}

func (testDateDomainModel) ModelName() string { return "test.date.domain" }

func (testDateDomainModel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "date_deadline", Type: Date},
		{Name: "name", Type: Char},
	}
}

func TestBuildSearchWhereClauseDateNotFalseIsNotNull(t *testing.T) {
	Registry["test.date.domain"] = testDateDomainModel{}
	t.Cleanup(func() { delete(Registry, "test.date.domain") })

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
