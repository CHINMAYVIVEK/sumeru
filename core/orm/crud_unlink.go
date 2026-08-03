package orm

import (
	"context"
	"fmt"

	"sumeru/core/applog"
)

// Unlink deletes a record by ID.
func Unlink(ctx context.Context, modelName string, id int) (err error) {
	defer func() { applog.ORMOp(ctx, "unlink", modelName, err, "id", id) }()
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
	if err == nil && !SecurityBypass(ctx) {
		AppendAudit(ctx, "unlink", modelName, int64(id), rec, nil, "")
	}
	return err
}
