package orm

import "fmt"

// FieldValidationError is a structured validation failure for a single model field.
type FieldValidationError struct {
	Field   string
	Label   string
	Message string
}

func (e *FieldValidationError) Error() string {
	if e == nil {
		return "validation error"
	}
	if e.Message != "" {
		return e.Message
	}
	label := e.Label
	if label == "" {
		label = e.Field
	}
	if label == "" {
		return "validation failed"
	}
	return fmt.Sprintf("%s is required.", label)
}

func newFieldValidationError(fd FieldDefinition, message string) error {
	label := fd.String
	if label == "" {
		label = fd.Name
	}
	if message == "" {
		message = fmt.Sprintf("%s is required.", label)
	}
	return &FieldValidationError{
		Field:   fd.Name,
		Label:   label,
		Message: message,
	}
}

// FieldValidationFields extracts field names from a validation error chain.
func FieldValidationFields(err error) []string {
	fve, ok := err.(*FieldValidationError)
	if !ok || fve == nil || fve.Field == "" {
		return nil
	}
	return []string{fve.Field}
}
