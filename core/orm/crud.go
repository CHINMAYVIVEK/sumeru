package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SearchInterceptor allows addons to intercept and modify search domains.
type SearchInterceptor func(ctx context.Context, model string, domain [][]interface{}) ([][]interface{}, error)

var (
	SearchInterceptors []SearchInterceptor
)

// RegisterSearchInterceptor adds an interceptor to the global ORM search pipeline.
func RegisterSearchInterceptor(fn SearchInterceptor) {
	SearchInterceptors = append(SearchInterceptors, fn)
}

// Create inserts a new record into the database
func Create(ctx context.Context, model Model, values map[string]interface{}) (int, error) {
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	if err := CheckRecordRules(ctx, uid, model.ModelName(), "create", values); err != nil {
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
	err := DB.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

// Upsert inserts or updates a record based on a unique field (usually 'name' or 'id')
func Upsert(ctx context.Context, model Model, values map[string]interface{}, conflictCol string) (int, error) {
	uid := SecurityUID(ctx)
	// UPSERT security: Check create/write permissions
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	// Note: Sumeru often skips record rules for upsert in bootstrap, but we should check them if possible.
	// For simplicity, we check model access. Full record rule check on upsert is complex because we don't know if it's create or update yet.

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

	// PostgreSQL requires at least one assignment in DO UPDATE SET; single-column upserts
	// (e.g. only conflict key) would otherwise produce "DO UPDATE SET RETURNING".
	if len(updates) == 0 {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", conflictCol, conflictCol))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING id",
		GetTableName(model.ModelName()), strings.Join(cols, ", "), strings.Join(placeholders, ", "),
		conflictCol, strings.Join(updates, ", "))

	var id int
	err := DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// Unlink deletes a record by ID.
func Unlink(ctx context.Context, modelName string, id int) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "unlink"); err != nil {
		return err
	}
	rec, err := SearchOne(ctx, modelName, map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	if err := CheckRecordRules(ctx, uid, modelName, "unlink", rec); err != nil {
		return err
	}

	tbl := GetTableName(modelName)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tbl)
	_, err = DB.ExecContext(ctx, query, id)
	return err
}

// SearchOne finds a single record matching the criteria
func SearchOne(ctx context.Context, modelName string, criteria map[string]interface{}) (map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not registered", modelName)
	}
	uid := SecurityUID(ctx)
	if !SecurityBypass(ctx) {
		if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
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
	rows, err := DB.QueryContext(ctx, query, args...)
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

	if !SecurityBypass(ctx) && uid > 0 {
		if err := CheckRecordRules(ctx, uid, modelName, "read", result); err != nil {
			return nil, sql.ErrNoRows
		}
	}

	return result, nil
}

// ResolveXmlId returns the database ID for a given XML ID (module.name).
// The name segment may contain dots (e.g. base.action_core.company → module base, name action_core.company).
func ResolveXmlId(ctx context.Context, xmlID string) (int, string, error) {
	parts := strings.Split(xmlID, ".")
	module := ""
	name := xmlID
	if len(parts) >= 2 {
		module = parts[0]
		name = strings.Join(parts[1:], ".")
	}

	criteria := map[string]interface{}{"name": name}
	if module != "" {
		criteria["module"] = module
	}

	data, err := SearchOne(ctx, "sys.model_data", criteria)
	if err != nil {
		return 0, "", err
	}
	rid, ok := CoerceInt64(data["core_id"])
	if !ok {
		return 0, "", fmt.Errorf("invalid core_id in sys.model_data")
	}
	return int(rid), AsString(data["model"]), nil
}

// Search finds records matching the criteria
func Search(ctx context.Context, modelName string, domain [][]interface{}) ([]map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}

	// Call registered interceptors (e.g. sumeru_ai)
	for _, interceptor := range SearchInterceptors {
		var err error
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return nil, err
		}
	}

	var err error
	domain, err = MergeRuleDomainsIntoSearch(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	whereClause, args, err := buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", GetTableName(modelName), whereClause)
	rows, err := DB.QueryContext(ctx, query, args...)
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
func SearchLimit(ctx context.Context, modelName string, domain [][]interface{}, limit int) ([]map[string]interface{}, error) {
	if _, ok := Registry[modelName]; !ok {
		return nil, fmt.Errorf("model %s not found", modelName)
	}
	if limit <= 0 {
		limit = 500
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "read"); err != nil {
		return nil, err
	}

	// Call registered interceptors
	for _, interceptor := range SearchInterceptors {
		var err error
		domain, err = interceptor(ctx, modelName, domain)
		if err != nil {
			return nil, err
		}
	}

	var err error
	domain, err = MergeRuleDomainsIntoSearch(ctx, uid, modelName, "read", domain)
	if err != nil {
		return nil, err
	}

	whereClause, args, err := buildSearchWhereClause(domain)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s ORDER BY id ASC LIMIT %d",
		GetTableName(modelName), whereClause, limit)
	rows, err := DB.QueryContext(ctx, query, args...)
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

// FindUIDefaultView returns the highest-priority sys.view for a model and type.
// view_mode "list" maps to type "tree".
func FindUIDefaultView(ctx context.Context, modelName, viewType string) (map[string]interface{}, error) {
	if _, ok := Registry["sys.view"]; !ok {
		return nil, fmt.Errorf("model sys.view not registered")
	}
	uid := SecurityUID(ctx)
	if !SecurityBypass(ctx) {
		if err := CheckModelAccess(ctx, uid, "sys.view", "read"); err != nil {
			return nil, err
		}
	}
	vt := strings.TrimSpace(strings.ToLower(viewType))
	if vt == "list" {
		vt = "tree"
	}
	tbl := GetTableName("sys.view")
	q := fmt.Sprintf(
		`SELECT * FROM %s WHERE model = $1 AND type = $2 ORDER BY priority DESC NULLS LAST, id DESC LIMIT 1`,
		tbl,
	)
	rows, err := DB.QueryContext(ctx, q, modelName, vt)
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
func UpdateRecordByID(ctx context.Context, modelName string, id int, values map[string]interface{}) error {
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "write"); err != nil {
		return err
	}
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return fmt.Errorf("model %s not found", modelName)
	}
	before, err := SearchOne(ctx, modelName, map[string]interface{}{"id": id})
	if err != nil {
		return err
	}
	if err := CheckRecordRules(ctx, uid, modelName, "write", before); err != nil {
		return err
	}

	// STAGE APPROVAL CHECK
	// If the model has a 'state' field and it's being changed, check for approval rules.
	if newState, ok := values["state"].(string); ok {
		oldState := AsString(before["state"])
		if newState != oldState {
			if err := CheckStageApproval(ctx, modelName, id, newState); err != nil {
				return err
			}
		}
	}

	merged := mergeRecordMap(before, values)
	if err := CheckRecordRules(ctx, uid, modelName, "write", merged); err != nil {
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
	_, execErr := DB.ExecContext(ctx, q, args...)
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

func quoteIdent(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
