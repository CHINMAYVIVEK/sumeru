package orm

import (
	"context"
	"fmt"
	"strings"
)

var Registry = map[string]Model{}

// modelDeclaringModule maps each technical model name to the declaring addon (sys.module.name).
// Kernel / platform models use "base" so metadata and DDL follow one module graph.
var modelDeclaringModule = make(map[string]string)

// RegisterModelWithModule registers a model and records which addon owns its table (catalog module linkage).
func RegisterModelWithModule(model Model, declaringModule string) {
	if model == nil {
		return
	}
	name := model.ModelName()
	Registry[name] = model
	modelDeclaringModule[name] = strings.TrimSpace(declaringModule)
}

func GetTableName(modelName string) string {
	return strings.ReplaceAll(modelName, ".", "_")
}

func SyncModels() error {
	if DB == nil {
		return nil
	}
	ctx := ContextWithBypass(context.Background(), true)
	installed, err := InstalledModuleNames(ctx)
	if err != nil {
		return err
	}
	for _, model := range Registry {
		name := model.ModelName()
		if len(installed) == 0 {
			owner := DeclaringModule(name)
			if owner != "" && !IsPlatformModule(owner) {
				continue
			}
		} else if !ShouldMaterializeModel(name, installed) {
			continue
		}
		if err := createTable(model); err != nil {
			return err
		}
	}
	return nil
}

// ColumnTypeSQL returns the PostgreSQL column type fragment for f (without NOT NULL / UNIQUE / DEFAULT).
// Many2One is stored as BIGINT. Unknown types return ok == false.
func ColumnTypeSQL(f FieldDefinition) (string, bool) {
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

		// Handle Default values
		defVal := field.DefaultVal
		if defVal != nil {
			switch v := defVal.(type) {
			case string:
				colType += fmt.Sprintf(" DEFAULT '%s'", strings.ReplaceAll(v, "'", "''"))
			case bool:
				if v {
					colType += " DEFAULT TRUE"
				} else {
					colType += " DEFAULT FALSE"
				}
			case int, int64, float64:
				colType += fmt.Sprintf(" DEFAULT %v", v)
			}
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
