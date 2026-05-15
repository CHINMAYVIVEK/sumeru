package orm

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
)

// UpdateRecordByID sets only columns that appear in the model's field definitions (never id).
func UpdateRecordByID(ctx context.Context, modelName string, id int, values map[string]interface{}) (err error) {
	defer func() { applog.ORMOp(ctx, "write", modelName, err, "id", id) }()
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
	_, err = DB.ExecContext(ctx, q, args...)
	return err
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
