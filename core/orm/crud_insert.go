package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sumeru/core/event"
)

// Create inserts a new record into the database using the shared values pipeline.
func Create(ctx context.Context, model Model, values map[string]interface{}) (id int, err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "create", model.ModelName(), err, "id", id) }()
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	prepared, err := PrepareValues(model, values, WriteOpCreate, PrepareOptions{StrictUnknown: true})
	if err != nil {
		return 0, err
	}
	if err := CheckFieldWriteAccess(ctx, uid, model.ModelName(), prepared); err != nil {
		return 0, err
	}
	if err := CheckRecordRules(ctx, uid, model.ModelName(), "create", prepared); err != nil {
		return 0, err
	}
	if len(prepared) == 0 {
		return 0, fmt.Errorf("create requires at least one column")
	}

	table, err := QuotedTableForModel(model.ModelName())
	if err != nil {
		return 0, err
	}
	var cols []string
	var placeholders []string
	var args []interface{}
	i := 1
	for col, val := range prepared {
		qcol, err := QuotedColumnForModel(model.ModelName(), col)
		if err != nil {
			return 0, err
		}
		cols = append(cols, qcol)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	err = DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == nil && !SecurityBypass(ctx) && !skipAuditModel(model.ModelName()) {
		AppendAudit(ctx, "create", model.ModelName(), int64(id), nil, prepared, "")
		EnqueueOutbox(ctx, "record.created", uid, map[string]interface{}{"model": model.ModelName(), "id": id})
		_ = event.Publish(ctx, event.Event{
			Name:    "record.created",
			Actor:   uid,
			Payload: map[string]interface{}{"model": model.ModelName(), "id": id},
		})
	}
	return id, err
}

// Upsert inserts or updates a record based on a unique field (usually 'name' or 'id').
// When SecurityBypass is set (module install), only field whitelist applies.
// Normal callers run field ACL and record rules on the write path when a row already exists.
func Upsert(ctx context.Context, model Model, values map[string]interface{}, conflictCol string) (id int, err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "upsert", model.ModelName(), err, "id", id) }()
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, model.ModelName(), "create"); err != nil {
		return 0, err
	}
	prepared, err := PrepareValues(model, values, WriteOpCreate, PrepareOptions{StrictUnknown: !SecurityBypass(ctx)})
	if err != nil {
		return 0, err
	}
	if !SecurityBypass(ctx) {
		if err := CheckFieldWriteAccess(ctx, uid, model.ModelName(), prepared); err != nil {
			return 0, err
		}
	}
	if conflictCol != "" {
		if _, ok := prepared[conflictCol]; !ok {
			if v, ok := values[conflictCol]; ok {
				prepared[conflictCol] = v
			}
		}
	}
	if len(prepared) == 0 {
		return 0, fmt.Errorf("upsert requires at least one column")
	}

	table, err := QuotedTableForModel(model.ModelName())
	if err != nil {
		return 0, err
	}
	conflictQuoted, err := QuotedConflictColumn(model.ModelName(), conflictCol)
	if err != nil {
		return 0, err
	}

	var cols []string
	var placeholders []string
	var updates []string
	var args []interface{}
	i := 1
	for col, val := range prepared {
		qcol, err := QuotedColumnForModel(model.ModelName(), col)
		if err != nil {
			return 0, err
		}
		cols = append(cols, qcol)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		args = append(args, val)
		if col != conflictCol {
			updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", qcol, qcol))
		}
		i++
	}
	if len(updates) == 0 {
		updates = append(updates, fmt.Sprintf("%s = EXCLUDED.%s", conflictQuoted, conflictQuoted))
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) ON CONFLICT (%s) DO UPDATE SET %s RETURNING id",
		table, strings.Join(cols, ", "), strings.Join(placeholders, ", "),
		conflictQuoted, strings.Join(updates, ", "))

	err = DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err == nil && SecurityBypass(ctx) {
		AppendAudit(ctx, "upsert", model.ModelName(), int64(id), nil, prepared, "source=module_sync")
	}
	return id, err
}
