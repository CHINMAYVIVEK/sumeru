package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// FindUIDefaultView returns the highest-priority sys.view for a model and type.
// view_mode "list" maps to type "tree".
func FindUIDefaultView(ctx context.Context, modelName, viewType string) (result map[string]interface{}, err error) {
	start := time.Now()
	defer func() {
		logORMOperation(ctx, start, "find_ui_view", "sys.view", err, "target_model", modelName, "view_type", viewType, "found", result != nil)
	}()
	if _, ok := Registry["sys.view"]; !ok {
		return nil, fmt.Errorf("model sys.view not registered")
	}
	if !SecurityBypass(ctx) {
		if err := CheckModelAccess(ctx, SecurityUID(ctx), "sys.view", "read"); err != nil {
			return nil, err
		}
	}
	vt := strings.TrimSpace(strings.ToLower(viewType))
	if vt == "list" {
		vt = "tree"
	}
	tbl := MustQuotedTableName("sys.view")
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
