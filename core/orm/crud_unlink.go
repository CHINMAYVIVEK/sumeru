package orm

import (
	"context"
	"fmt"
	"time"

	"sumeru/core/event"
)

// Unlink deletes a record by ID using domain-scoped UnlinkWhere.
func Unlink(ctx context.Context, modelName string, id int) (err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "unlink", modelName, err, "id", id) }()
	if id <= 0 {
		return fmt.Errorf("invalid id")
	}
	_, err = UnlinkWhere(ctx, modelName, [][]interface{}{{"id", "=", id}})
	return err
}

// UnlinkWhere deletes rows matching domain with write/unlink record-rule predicates
// applied in SQL inside a transaction with FOR UPDATE.
func UnlinkWhere(ctx context.Context, modelName string, domain [][]interface{}) (n int64, err error) {
	start := time.Now()
	defer func() { logORMOperation(ctx, start, "unlink_where", modelName, err, "rows", n) }()
	uid := SecurityUID(ctx)
	if err := CheckModelAccess(ctx, uid, modelName, "unlink"); err != nil {
		return 0, err
	}
	if _, ok := Registry[modelName]; !ok {
		return 0, fmt.Errorf("model %s not found", modelName)
	}

	table, err := QuotedTableForModel(modelName)
	if err != nil {
		return 0, err
	}
	securedSQL, args, err := BuildWhereWithRecordRules(ctx, uid, modelName, "unlink", domain)
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
	for _, rec := range locked {
		if err := CheckRecordRules(ctx, uid, modelName, "unlink", rec); err != nil {
			return 0, err
		}
	}

	delQ := fmt.Sprintf(`DELETE FROM %s WHERE %s`, table, securedSQL)
	res, err := tx.ExecContext(ctx, delQ, args...)
	if err != nil {
		return 0, err
	}
	n, _ = res.RowsAffected()
	if n == 0 {
		return 0, fmt.Errorf("access denied or record not found")
	}

	if !SecurityBypass(ctx) {
		for _, rec := range locked {
			rid, _ := CoerceInt64(rec["id"])
			AppendAudit(ctx, "unlink", modelName, rid, rec, nil, "")
			EnqueueOutbox(ctx, "record.deleted", uid, map[string]interface{}{"model": modelName, "id": int(rid)})
			_ = event.Publish(ctx, event.Event{
				Name:    "record.deleted",
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
