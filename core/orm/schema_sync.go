package orm

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

// SyncRegistrySchema adds missing columns and indexes for every model in Registry
// by comparing FieldDefinition to information_schema (PostgreSQL). Safe on populated tables:
// new columns are nullable unless a DEFAULT is declared (booleans, defaults).
//
// Call after SyncModels() on startup, and on module install (-i) / update (-u) so schema
// tracks Go model definitions without hand-written ALTER lists.
func SyncRegistrySchema() error {
	if DB == nil {
		return nil
	}
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		m := Registry[name]
		if err := syncModelSchema(m); err != nil {
			return fmt.Errorf("schema sync %s: %w", name, err)
		}
	}
	return nil
}

func syncModelSchema(model Model) error {
	tableName := GetTableName(model.ModelName())
	exists, err := tableExists(tableName)
	if err != nil {
		return err
	}
	if !exists {
		return createTable(model)
	}
	existing, err := loadTableColumns(tableName)
	if err != nil {
		return err
	}
	for _, field := range model.Fields() {
		if field.Name == "id" {
			continue
		}
		if _, ok := existing[strings.ToLower(field.Name)]; ok {
			continue
		}
		baseType, ok := ColumnTypeSQL(field)
		if !ok {
			continue
		}
		colDef := buildAddColumnDefinition(field, baseType)
		q := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", quoteIdent(tableName), quoteIdent(field.Name), colDef)
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
		log.Printf("schema sync: %s.%s added", tableName, field.Name)
	}
	return ensureModelIndexes(tableName, model)
}

func ensureModelIndexes(tableName string, model Model) error {
	for _, field := range model.Fields() {
		if !(field.Index || field.Type == Many2One) {
			continue
		}
		idxName := fmt.Sprintf("idx_%s_%s", tableName, field.Name)
		idxQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", quoteIdent(idxName), quoteIdent(tableName), quoteIdent(field.Name))
		if _, err := DB.Exec(idxQuery + ";"); err != nil {
			return fmt.Errorf("index %s: %w", idxName, err)
		}
	}
	return nil
}

func tableExists(tableName string) (bool, error) {
	var n int
	err := DB.QueryRow(`
		SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func loadTableColumns(tableName string) (map[string]struct{}, error) {
	rows, err := DB.Query(`
		SELECT column_name FROM information_schema.columns
		WHERE table_schema = 'public' AND table_name = $1
	`, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string]struct{})
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		out[strings.ToLower(c)] = struct{}{}
	}
	return out, rows.Err()
}


func buildAddColumnDefinition(f FieldDefinition, baseType string) string {
	if f.Type == Boolean {
		if f.DefaultVal == true {
			return baseType + " NOT NULL DEFAULT TRUE"
		}
		if f.DefaultVal == false {
			return baseType + " NOT NULL DEFAULT FALSE"
		}
		if lit, ok := sqlDefaultLiteral(f.DefaultVal); ok {
			return baseType + " DEFAULT " + lit
		}
		return baseType
	}
	if lit, ok := sqlDefaultLiteral(f.DefaultVal); ok {
		if f.Required {
			return baseType + " NOT NULL DEFAULT " + lit
		}
		return baseType + " DEFAULT " + lit
	}
	return baseType
}

func sqlDefaultLiteral(v interface{}) (string, bool) {
	if v == nil {
		return "", false
	}
	switch t := v.(type) {
	case bool:
		if t {
			return "TRUE", true
		}
		return "FALSE", true
	case int:
		return fmt.Sprintf("%d", t), true
	case int64:
		return fmt.Sprintf("%d", t), true
	case float32:
		return fmt.Sprintf("%g", float64(t)), true
	case float64:
		return fmt.Sprintf("%g", t), true
	case string:
		return "'" + strings.ReplaceAll(t, "'", "''") + "'", true
	default:
		return "'" + strings.ReplaceAll(fmt.Sprint(t), "'", "''") + "'", true
	}
}
