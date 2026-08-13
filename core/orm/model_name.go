package orm

import (
	"fmt"
	"regexp"
	"strings"
)

// Logical model names: one or more lowercase segments joined only by '.'.
// Each segment is [a-z][a-z0-9]* — no underscore, whitespace, or other symbols.
var modelNameSegment = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// Field / column names may use underscores: [a-z][a-z0-9_]*
var fieldNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ValidateModelName returns an error if name is not a legal logical model name.
func ValidateModelName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("model name is empty")
	}
	if strings.ContainsAny(name, " \t\n\r") {
		return fmt.Errorf("model name %q must not contain whitespace", name)
	}
	if strings.Contains(name, "_") {
		return fmt.Errorf("model name %q must not contain underscore; use '.' to join segments", name)
	}
	parts := strings.Split(name, ".")
	for i, p := range parts {
		if p == "" {
			return fmt.Errorf("model name %q has empty segment at position %d", name, i)
		}
		if !modelNameSegment.MatchString(p) {
			return fmt.Errorf("model name segment %q in %q is invalid (expect [a-z][a-z0-9]*)", p, name)
		}
	}
	return nil
}

// ValidateFieldName returns an error if field is not a legal SQL column identifier.
func ValidateFieldName(field string) error {
	field = strings.TrimSpace(field)
	if field == "" {
		return fmt.Errorf("field name is empty")
	}
	if !fieldNamePattern.MatchString(field) {
		return fmt.Errorf("field name %q is invalid (expect [a-z][a-z0-9_]*)", field)
	}
	return nil
}

// ModelToTableName validates name, then returns the physical table name (dots → underscores).
// Use for information_schema, tableExists, and index/constraint name prefixes. Never pass
// the result through quoteIdent twice.
func ModelToTableName(name string) (string, error) {
	if err := ValidateModelName(name); err != nil {
		return "", err
	}
	return strings.ReplaceAll(name, ".", "_"), nil
}

// QuotedTableName validates and returns a quoted physical table name for SQL clauses
// (FROM, UPDATE, ALTER TABLE, ON). Never concatenate into another identifier (e.g. idx_ + quoted).
func QuotedTableName(name string) (string, error) {
	table, err := ModelToTableName(name)
	if err != nil {
		return "", err
	}
	return quoteIdent(table), nil
}

// MustModelToTableName is for known-good compile-time literals; panics on invalid names.
// Physical name only: catalog lookups and index names, not SQL clauses.
func MustModelToTableName(name string) string {
	table, err := ModelToTableName(name)
	if err != nil {
		panic(err)
	}
	return table
}

// MustQuotedTableName is for known-good compile-time literals; panics on invalid names.
// SQL clauses only; never splice into idx_ / constraint names.
func MustQuotedTableName(name string) string {
	q, err := QuotedTableName(name)
	if err != nil {
		panic(err)
	}
	return q
}
