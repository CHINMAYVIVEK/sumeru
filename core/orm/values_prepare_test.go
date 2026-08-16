package orm

import "testing"

type testDateModel struct {
	BaseModel
}

func (testDateModel) ModelName() string { return "test.date" }

func (testDateModel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "date_deadline", Type: Date, String: "Deadline"},
		{Name: "date_last_stage_update", Type: DateTime, String: "Last Stage Update"},
	}
}

func TestCoerceFieldValueEmptyDateBecomesNil(t *testing.T) {
	fd := FieldDefinition{Name: "date_deadline", Type: Date}
	got, err := coerceFieldValue(fd, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty date, got %#v", got)
	}
}

func TestCoerceFieldValueEmptyDateTimeBecomesNil(t *testing.T) {
	fd := FieldDefinition{Name: "date_last_stage_update", Type: DateTime}
	got, err := coerceFieldValue(fd, "   ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for empty datetime, got %#v", got)
	}
}

func TestCoerceFieldValueNonEmptyDatePreserved(t *testing.T) {
	fd := FieldDefinition{Name: "date_deadline", Type: Date}
	got, err := coerceFieldValue(fd, "2026-08-15")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "2026-08-15" {
		t.Fatalf("expected date string preserved, got %#v", got)
	}
}

func TestPrepareValuesEmptyOptionalDateIsNil(t *testing.T) {
	out, err := PrepareValues(testDateModel{}, map[string]interface{}{
		"name":           "Acme",
		"date_deadline":  "",
	}, WriteOpCreate, PrepareOptions{StrictUnknown: true})
	if err != nil {
		t.Fatalf("PrepareValues: %v", err)
	}
	if out["date_deadline"] != nil {
		t.Fatalf("expected nil date_deadline, got %#v", out["date_deadline"])
	}
}

func TestPrepareValuesRequiredFieldValidationError(t *testing.T) {
	_, err := PrepareValues(testDateModel{}, map[string]interface{}{}, WriteOpCreate, PrepareOptions{StrictUnknown: true})
	fve, ok := err.(*FieldValidationError)
	if !ok {
		t.Fatalf("expected FieldValidationError, got %T: %v", err, err)
	}
	if fve.Field != "name" || fve.Label != "Name" {
		t.Fatalf("unexpected validation error: %+v", fve)
	}
}
