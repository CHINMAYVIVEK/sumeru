package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sumeru/core/applog"
)

// SearchOne finds a single record matching the criteria
func SearchOne(ctx context.Context, modelName string, criteria map[string]interface{}) (result map[string]interface{}, err error) {
	defer func() {
		applog.ORMOp(ctx, "search_one", modelName, err, "has_row", result != nil)
	}()
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
		err = sql.ErrNoRows
		return nil, err
	}

	cols, _ := rows.Columns()
	vals := make([]interface{}, len(cols))
	valPtrs := make([]interface{}, len(cols))
	for i := range vals {
		valPtrs[i] = &vals[i]
	}

	if err = rows.Scan(valPtrs...); err != nil {
		return nil, err
	}

	result = make(map[string]interface{})
	for i, col := range cols {
		result[col] = vals[i]
	}

	if !SecurityBypass(ctx) && uid > 0 {
		if e := CheckRecordRules(ctx, uid, modelName, "read", result); e != nil {
			err = sql.ErrNoRows
			return nil, err
		}
	}

	return result, nil
}
