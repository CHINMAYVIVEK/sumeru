package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// FindUIDefaultView returns the highest-priority sys.view for a model and type.
// view_mode "list" maps to type "tree". Record rules are applied in SQL like Search.
func FindUIDefaultView(ctx context.Context, modelName, viewType string) (result map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		logORMOperationKV(ctx, start, "find_ui_view", "sys.view", err, "target_model", modelName, "view_type", viewType, "found", result != nil)
	}()
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
	domain := [][]interface{}{
		{"model", "=", modelName},
		{"type", "=", vt},
	}
	whereClause, args, err := BuildWhereWithRecordRules(ctx, uid, "sys.view", "read", domain)
	if err != nil {
		return nil, err
	}
	priCol, err := QuotedColumnForModel("sys.view", "priority")
	if err != nil {
		return nil, err
	}
	idCol, err := QuotedColumnForModel("sys.view", "id")
	if err != nil {
		return nil, err
	}
	tbl, err := QuotedTableForModel("sys.view")
	if err != nil {
		return nil, err
	}
	q := fmt.Sprintf(
		`SELECT * FROM %s WHERE %s ORDER BY %s DESC NULLS LAST, %s DESC LIMIT 1`,
		tbl, whereClause, priCol, idCol,
	)
	rows, err := DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		err = sql.ErrNoRows
		return nil, err
	}

	cols, _ := rows.Columns()
	result, err = scanRowToMap(cols, rows)
	if err != nil {
		return nil, err
	}
	return result, nil
}
