package orm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/event"
)

// UpdateRecordByID sets columns from prepared values for a single id.
// Security uses the same domain-scoped Update path (record-rule SQL + row lock).
func UpdateRecordByID(ctx context.Context, modelName string, id int, values map[string]interface{}) (err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "write", modelName, err, "id", id) }()
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err = Update(ctx, modelName, [][]interface{}{{"id", "=", id}}, values)
	return err
}

// Update updates rows matching domain with values. Record rules for write are
// compiled into the WHERE clause; mutations run inside a transaction with FOR UPDATE.
// Returns rows affected. Zero rows → access denied or not found.
func Update(ctx context.Context, modelName string, domain [][]interface{}, values map[string]interface{}) (n int64, err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "update", modelName, err, "rows", n) }()
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "write"); err != nil {
		return 0, err
	}
	inst, ok := Registry[modelName]
	if !ok || inst == nil {
		return 0, fmt.Errorf("model %s not found", modelName)
	}
	prepared, err := PrepareValues(inst, values, WriteOpWrite, PrepareOptions{StrictUnknown: false})
	if err != nil {
		return 0, err
	}
	if err := CheckFieldWriteAccess(ctx, uid, modelName, prepared); err != nil {
		return 0, err
	}
	if len(prepared) == 0 {
		return 0, nil
	}

	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return 0, err
	}
	securedSQL, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "write", domain)
	if err != nil {
		return 0, err
	}

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback() }()

	lockQ := fmt.Sprintf(`SELECT * FROM %s WHERE %s FOR UPDATE`, table, securedSQL)
	rows, err := tx.QueryContext(ctx, lockQ, args...)
	if err != nil {
		return 0, err
	}
	cols, _ := rows.Columns()
	var locked []map[string]interface{}
	for rows.Next() {
		m, err := scanRowToMap(cols, rows)
		if err != nil {
			rows.Close()
			return 0, err
		}
		locked = append(locked, m)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	if len(locked) == 0 {
		return 0, fmt.Errorf("access denied or record not found")
	}

	for _, before := range locked {
		if newState, ok := prepared["state"].(string); ok {
			oldState := AsString(before["state"])
			if newState != oldState {
				rid, _ := CoerceInt64(before["id"])
				if err := CanWorkflowTransition(ctx, modelName, int(rid), oldState, newState, uid); err != nil {
					return 0, err
				}
			}
		}
		merged := mergeRecordMap(before, prepared)
		if err := CheckRecordRules(ctx, uid, modelName, "write", merged); err != nil {
			return 0, err
		}
	}

	var sets []string
	var setArgs []interface{}
	i := 1
	for k, v := range prepared {
		qcol, err := QuotedColumnForModel(modelName, k)
		if err != nil {
			return 0, err
		}
		sets = append(sets, fmt.Sprintf("%s = $%d", qcol, i))
		setArgs = append(setArgs, v)
		i++
	}
	shiftedWhere, _ := shiftPlaceholders(securedSQL, len(setArgs)+1)
	allArgs := append(setArgs, args...)
	updQ := fmt.Sprintf(`UPDATE %s SET %s WHERE %s`, table, strings.Join(sets, ", "), shiftedWhere)
	res, err := tx.ExecContext(ctx, updQ, allArgs...)
	if err != nil {
		return 0, err
	}
	n, _ = res.RowsAffected()
	if n == 0 {
		return 0, fmt.Errorf("access denied or record not found")
	}

	if !SecurityBypass(ctx) {
		for _, before := range locked {
			rid, _ := CoerceInt64(before["id"])
			merged := mergeRecordMap(before, prepared)
			AppendAudit(ctx, "write", modelName, rid, before, merged, "")
			EnqueueOutbox(ctx, "record.updated", uid, map[string]interface{}{"model": modelName, "id": int(rid)})
			_ = event.Publish(ctx, event.Event{
				Name:    "record.updated",
				Actor:   uid,
				Payload: map[string]interface{}{"model": modelName, "id": int(rid)},
			})
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return n, nil
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
