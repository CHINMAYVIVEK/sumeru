package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

// Create inserts a new record into the database
func Create(model Model, values map[string]interface{}) (int, error) {
	uid := SecurityUID()
	if err := CheckModelAccess(uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	if err := CheckRecordRules(uid, model.ModelName(), "create", values); err != nil {
		return 0, err
	}
	var cols []string
	var placeholders []string
	var args []interface{}

	i := 1
	for col, val := range values {
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id", 
		GetTableName(model.ModelName()), strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	var id int
	err := DB.QueryRow(query, args...).Scan(&id)
	return id, err
}

// Upsert inserts or updates a record based on a unique field (usually 'name' or 'id')
func Upsert(model Model, values map[string]interface{}, conflictCol string) (int, error) {
	var cols []string
	var placeholders []string
	var updates []string
	var args []interface{}

	i := 1
	for col, val := range values {
		cols = append(cols, col)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		if col != conflictCol {
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", col, col))
		}
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING id",
		GetTableName(model.ModelName()), strings.Join(cols, ", "), strings.Join(placeholders, ", "), 
		conflictCol, strings.Join(updates, ", "))

	var id int
	err := DB.QueryRow(query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// SearchOne finds a single record matching the criteria
func SearchOne(modelName string, criteria map[string]interface{}) (map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not registered", modelName)
	}
	uid := SecurityUID()
	if !SecurityBypass() {
		if err := CheckModelAccess(uid, modelName, "read"); err != nil {
			return nil, err
		}
	}

	var where []string
	var args []interface{}
	i := 1
	for col, val := range criteria {
		where = append(where, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", GetTableName(modelName), strings.Join(where, " AND "))
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}

	if err := rows.Scan(valPtrs...); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for i, col := range cols {
		result[col] = vals[i]
	}

	if !SecurityBypass() && uid > 0 {
		if err := CheckRecordRules(uid, modelName, "read", result); err != nil {
			return nil, sql.ErrNoRows
		}
	}

	return result, nil
}

// ResolveXmlId returns the database ID for a given XML ID (module.name)
func ResolveXmlId(xmlID string) (int, string, error) {
	parts := strings.Split(xmlID, ".")
	module := ""
	name := xmlID
	if len(parts) == 2 {
		module = parts[0]
		name = parts[1]
	}

	criteria := map[string]interface{}{"name": name}
	if module != "" {
		criteria["module"] = module
	}

	data, err := SearchOne("ir.model.data", criteria)
	if err != nil {
		return 0, "", err
	}
	rid, ok := CoerceInt64(data["res_id"])
	if !ok {
		return 0, "", fmt.Errorf("invalid res_id in ir.model.data")
	}
	return int(rid), AsString(data["model"]), nil
}

// Search finds records matching the criteria
func Search(modelName string, domain [][]interface{}) ([]map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID()
	if err := CheckModelAccess(uid, modelName, "read"); err != nil {
		return nil, err
	}
	var err error
	domain, err = MergeRuleDomainsIntoSearch(uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	whereClause, args, err := buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", GetTableName(modelName), whereClause)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	results := []map[string]interface{}{}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}

		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}

		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = vals[i]
		}
		results = append(results, m)
	}

	return results, rows.Err()
}

// SearchLimit returns up to limit rows for modelName matching domain, ordered by id.
// limit must be positive; otherwise it defaults to 500.
func SearchLimit(modelName string, domain [][]interface{}, limit int) ([]map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	if limit <= 0 {
		limit = 500
	}
	uid := SecurityUID()
	if err := CheckModelAccess(uid, modelName, "read"); err != nil {
		return nil, err
	}
	var err error
	domain, err = MergeRuleDomainsIntoSearch(uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	whereClause, args, err := buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY id ASC LIMIT %d",
		GetTableName(modelName), whereClause, limit)
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	var results []map[string]interface{}

	for rows.Next() {
		vals := make([]interface{}, len(cols))
		valPtrs := make([]interface{}, len(cols))
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for i, col := range cols {
			m[col] = vals[i]
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

// CoerceInt64 reads numeric values from database drivers into int64.
func CoerceInt64(v interface{}) (int64, bool) {
	switch t := v.(type) {
	case int64:
		return t, true
	case int32:
		return int64(t), true
	case int:
		return int64(t), true
	case uint64:
		return int64(t), true
	case uint32:
		return int64(t), true
	case float64:
		return int64(t), true
	case float32:
		return int64(t), true
	case string:
		s := strings.TrimSpace(t)
		if s == "" {
			return 0, false
		}
		var n int64
		_, err := fmt.Sscanf(s, "%d", &n)
		return n, err == nil
	case []byte:
		var n int64
		_, err := fmt.Sscanf(string(t), "%d", &n)
		return n, err == nil
	default:
		return 0, false
	}
}

// AsString coerces common database driver values to string.
func AsString(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

// FindUIDefaultView returns the highest-priority ir.ui.view for a model and type.
// view_mode "list" maps to type "tree".
func FindUIDefaultView(modelName, viewType string) (map[string]interface{}, error) {
	if _, ok := Registry["ir.ui.view"]; !ok {
		return nil, fmt.Errorf("model ir.ui.view not registered")
	}
	uid := SecurityUID()
	if !SecurityBypass() {
		if err := CheckModelAccess(uid, "ir.ui.view", "read"); err != nil {
			return nil, err
		}
	}
	vt := strings.TrimSpace(strings.ToLower(viewType))
	if vt == "list" {
		vt = "tree"
	}
	tbl := GetTableName("ir.ui.view")
	q := fmt.Sprintf(
		`SELECT * FROM %s WHERE model = $1 AND type = $2 ORDER BY priority DESC NULLS LAST, id DESC LIMIT 1`,
		tbl,
	)
	rows, err := DB.Query(q, modelName, vt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}
	if err := rows.Scan(valPtrs...); err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for i, col := range cols {
		result[col] = vals[i]
	}
	return result, nil
}

// UpdateRecordByID sets only columns that appear in the model's field definitions (never id).
func UpdateRecordByID(modelName string, id int, values map[string]interface{}) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	uid := SecurityUID()
	if err := CheckModelAccess(uid, modelName, "write"); err != nil {
		return err
	}
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return fmt.Errorf("model %s not found", modelName)
	}
	before, err := SearchOne(modelName, map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	if err := CheckRecordRules(uid, modelName, "write", before); err != nil {
		return err
	}
	merged := mergeRecordMap(before, values)
	if err := CheckRecordRules(uid, modelName, "write", merged); err != nil {
		return err
	}
	allowed := map[string]struct{}{}
	for _, f := range inst.Fields() {
		if f.Name != "" && f.Name != "id" {
			allowed[f.Name] = struct{}{}
		}
	}
	var sets []string
	var args []interface{}
	i := 1
	for k, v := range values {
		if k == "id" {
			continue
		}
		if _, ok := allowed[k]; !ok {
			continue
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", k, i))
		args = append(args, v)
		i++
	}
	if len(sets) == 0 {
		return nil
	}
	args = append(args, id)
	tbl := GetTableName(modelName)
	q := fmt.Sprintf(`UPDATE %s SET %s WHERE id = $%d`, tbl, strings.Join(sets, ", "), i)
	_, execErr := DB.Exec(q, args...)
	return execErr
}

func mergeRecordMap(base map[string]interface{}, patch map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for k, v := range base {
		out[k] = v
	}
	for k, v := range patch {
		out[k] = v
	}
	return out
}

func buildSearchWhereClause(domain [][]interface{}) (string, []interface{}, error) {
	if len(domain) == 0 {
		return "1=1", nil, nil
	}
	var parts []string
	var args []interface{}
	n := 1
	for _, d := range domain {
		if len(d) != 3 {
			return "", nil, fmt.Errorf("invalid domain clause %v", d)
		}
		field, ok := d[0].(string)
		if !ok || strings.TrimSpace(field) == "" {
			return "", nil, fmt.Errorf("domain field name")
		}
		op := strings.TrimSpace(strings.ToLower(fmt.Sprint(d[1])))
		col := quoteIdent(field)
		switch op {
		case "=":
			parts = append(parts, fmt.Sprintf("%s = $%d", col, n))
			args = append(args, d[2])
			n++
		case "!=":
			parts = append(parts, fmt.Sprintf("(%s IS DISTINCT FROM $%d)", col, n))
			args = append(args, d[2])
			n++
		case "in":
			list, ok := d[2].([]interface{})
			if !ok {
				return "", nil, fmt.Errorf("operator in requires array value")
			}
			if len(list) == 0 {
				parts = append(parts, "FALSE")
				continue
			}
			ph := make([]string, len(list))
			for i := range list {
				ph[i] = fmt.Sprintf("$%d", n)
				args = append(args, list[i])
				n++
			}
			parts = append(parts, fmt.Sprintf("%s IN (%s)", col, strings.Join(ph, ",")))
		default:
			return "", nil, fmt.Errorf("unsupported domain operator %q", op)
		}
	}
	return strings.Join(parts, " AND "), args, nil
}
