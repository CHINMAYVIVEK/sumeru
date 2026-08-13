package orm

import (
	"fmt"
	"strings"
)

// QuotedTableForModel resolves a registered model to a quoted physical table name.
func QuotedTableForModel(modelName string) (string, error) {
	modelName = strings.TrimSpace(modelName)
	if _, ok := Registry[modelName]; !ok {
		return "", fmt.Errorf("model %q not registered", modelName)
	}
	return QuotedTableName(modelName)
}

// QuotedColumnForModel returns a quoted column identifier if field is id or a
// declared SQL column on the model (excludes Many2Many / One2Many).
func QuotedColumnForModel(modelName, fieldName string) (string, error) {
	modelName = strings.TrimSpace(modelName)
	fieldName = strings.TrimSpace(fieldName)
	m, ok := Registry[modelName]
	if !ok || m == nil {
		return "", fmt.Errorf("model %q not registered", modelName)
	}
	if err := ValidateFieldName(fieldName); err != nil {
		return "", err
	}
	if fieldName == "id" {
		return quoteIdent("id"), nil
	}
	for _, f := range m.Fields() {
		if f.Name != fieldName {
			continue
		}
		if f.Type == Many2Many || f.Type == One2Many {
			return "", fmt.Errorf("field %q on %s is relational and not a SQL column", fieldName, modelName)
		}
		return quoteIdent(fieldName), nil
	}
	return "", fmt.Errorf("unknown field %q on model %s", fieldName, modelName)
}

// isSQLColumnOnModel reports whether fieldName is a writable/searchable SQL column.
func isSQLColumnOnModel(modelName, fieldName string) bool {
	_, err := QuotedColumnForModel(modelName, fieldName)
	return err == nil
}

// QuotedConflictColumn validates an Upsert conflict target column.
func QuotedConflictColumn(modelName, column string) (string, error) {
	return QuotedColumnForModel(modelName, column)
}

// ParseOrderByForModel parses "field", "field ASC", or "field DESC" against the model.
// Empty orderBy defaults to "id ASC". Free-form SQL is rejected.
func ParseOrderByForModel(modelName, orderBy string) (string, error) {
	ob := strings.TrimSpace(orderBy)
	if ob == "" {
		col, err := QuotedColumnForModel(modelName, "id")
		if err != nil {
			return "", err
		}
		return col + " ASC", nil
	}
	parts := strings.Fields(ob)
	if len(parts) == 0 || len(parts) > 2 {
		return "", fmt.Errorf("invalid order by %q", orderBy)
	}
	field := parts[0]
	dir := "ASC"
	if len(parts) == 2 {
		d := strings.ToUpper(parts[1])
		if d != "ASC" && d != "DESC" {
			return "", fmt.Errorf("invalid order direction %q", parts[1])
		}
		dir = d
	}
	col, err := QuotedColumnForModel(modelName, field)
	if err != nil {
		return "", err
	}
	return col + " " + dir, nil
}
