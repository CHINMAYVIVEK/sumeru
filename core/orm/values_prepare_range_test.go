package orm

import "testing"

func TestPrepareValuesMinMaxValidation(t *testing.T) {
	m := stubModel{
		name: "test.range",
		fields: []FieldDefinition{
			{Name: "rating", Type: Integer, String: "Rating", Min: floatPtr(0), Max: floatPtr(5)},
		},
	}
	_, err := PrepareValues(m, map[string]interface{}{"rating": 10}, WriteOpCreate, PrepareOptions{StrictUnknown: true})
	if err == nil {
		t.Fatal("expected validation error for rating > max")
	}
	out, err := PrepareValues(m, map[string]interface{}{"rating": 3}, WriteOpCreate, PrepareOptions{StrictUnknown: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out["rating"] != 3 {
		t.Fatalf("got %#v", out["rating"])
	}
}

func floatPtr(f float64) *float64 { return &f }
