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
		err := createTable(model)
		if err != nil {
			return err
		}
	}
	return nil
}

func createTable(model Model) error {
	tableName := GetTableName(model.ModelName())
	var columns []string
	columns = append(columns, "id BIGSERIAL PRIMARY KEY")

	for _, field := range model.Fields() {
		colType := ""
		switch field.Type {
		case Char:
			colType = "VARCHAR(255)"
		case Text:
			colType = "TEXT"
		case Integer:
			colType = "BIGINT"
		case Float:
			colType = "DOUBLE PRECISION"
		case Numeric:
			colType = "NUMERIC(16, 4)" // 16 digits total, 4 decimal places
		case Boolean:
			colType = "BOOLEAN"
		case Date:
			colType = "DATE"
		case DateTime:
			colType = "TIMESTAMPTZ"
		case Selection:
			colType = "VARCHAR(50)"
		case Json:
			colType = "JSONB"
		case Many2One:
			colType = "BIGINT" // Reference to other table's BIGSERIAL ID
		default:
			continue
		}

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
