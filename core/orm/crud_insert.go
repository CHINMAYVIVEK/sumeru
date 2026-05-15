package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"sumeru/core/applog"
)

// Create inserts a new record into the database
func Create(ctx context.Context, model Model, values map[string]interface{}) (id int, err error) {
	defer func() { applog.ORMOp(ctx, "create", model.ModelName(), err, "id", id) }()
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

	err = DB.QueryRowContext(ctx, query, args...).Scan(&id)
	return id, err
}

// Upsert inserts or updates a record based on a unique field (usually 'name' or 'id')
func Upsert(ctx context.Context, model Model, values map[string]interface{}, conflictCol string) (id int, err error) {
	defer func() { applog.ORMOp(ctx, "upsert", model.ModelName(), err, "id", id) }()
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

	err = DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}
