package orm

import (
	"fmt"
	"strings"
)

var Registry = map[string]Model{}

func RegisterModel(model Model) {
	Registry[model.ModelName()] = model
}

func GetTableName(modelName string) string {
	return strings.ReplaceAll(modelName, ".", "_")
}

func SyncModels() error {
	for _, model := range Registry {
		if err := createTable(model); err != nil {
			return err
		}
	}
	return nil
}

// ColumnTypeSQL returns the PostgreSQL column type fragment for f (without NOT NULL / UNIQUE / DEFAULT).
// Many2One is stored as BIGINT. Unknown types return ok == false.
func ColumnTypeSQL(f FieldDefinition) (sql string, ok bool) {
	switch f.Type {
	case Char:
		return "VARCHAR(255)", true
	case Text:
		return "TEXT", true
	case Integer:
		return "BIGINT", true
	case Float:
		return "DOUBLE PRECISION", true
	case Numeric:
		return "NUMERIC(16, 4)", true
	case Boolean:
		return "BOOLEAN", true
	case Date:
		return "DATE", true
	case DateTime:
		return "TIMESTAMPTZ", true
	case Selection:
		return "VARCHAR(50)", true
	case Json:
		return "JSONB", true
	case Many2One:
		return "BIGINT", true
	default:
		return "", false
	}
}

func createTable(model Model) error {
	tableName := GetTableName(model.ModelName())
	var columns []string
	columns = append(columns, "id BIGSERIAL PRIMARY KEY")

	for _, field := range model.Fields() {
		baseType, ok := ColumnTypeSQL(field)
		if !ok {
			continue
		}
		colType := baseType
		if field.Required {
			colType += " NOT NULL"
		}
		if field.Unique {
			colType += " UNIQUE"
		}
		columns = append(columns, fmt.Sprintf("%s %s", field.Name, colType))
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, strings.Join(columns, ", "))
	fmt.Printf("Syncing model %s...\n", model.ModelName())

	if _, err := DB.Exec(query); err != nil {
		return err
	}

	// Handle Indexing
	for _, field := range model.Fields() {
		if field.Index || field.Type == Many2One {
			idxName := fmt.Sprintf("idx_%s_%s", tableName, field.Name)
			idxQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s);", idxName, tableName, field.Name)
			fmt.Printf("Syncing index %s...\n", idxName)
			if _, err := DB.Exec(idxQuery); err != nil {
				fmt.Printf("Error creating index %s: %v\n", idxName, err)
			}
		}
	}

	return nil
}
