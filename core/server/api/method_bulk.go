package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcCreateMany(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "create_many requires args[0] list of values", map[string]interface{}{"method": "create_many"})
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(arr[0], &rows); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] values list: %v", err), map[string]interface{}{"method": "create_many"})
	}
	inst, ok := orm.Registry[model]
	if !ok || inst == nil {
		return nil, newRPCError(CodeModelNotFound, fmt.Sprintf("model %q not registered", model), map[string]interface{}{"model": model})
	}
	ids := make([]int, 0, len(rows))
	for i, vals := range rows {
		id, err := orm.Create(ctx, inst, vals)
		if err != nil {
			return nil, fmt.Errorf("create_many[%d]: %w", i, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func rpcWriteMany(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "write_many requires args[0] ids and args[1] values", map[string]interface{}{"method": "write_many"})
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] ids: %v", err), map[string]interface{}{"method": "write_many"})
	}
	var vals map[string]interface{}
	if err := json.Unmarshal(arr[1], &vals); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] values: %v", err), map[string]interface{}{"method": "write_many"})
	}
	domain := make([][]interface{}, 0, len(ids))
	if len(ids) == 0 {
		return true, nil
	}
	idList := make([]interface{}, len(ids))
	for i, id := range ids {
		idList[i] = id
	}
	domain = append(domain, []interface{}{"id", "in", idList})
	_, err = orm.Update(ctx, model, domain, vals)
	if err != nil {
		return nil, err
	}
	return true, nil
}

func rpcUnlinkMany(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "unlink_many requires args[0] ids", map[string]interface{}{"method": "unlink_many"})
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] ids: %v", err), map[string]interface{}{"method": "unlink_many"})
	}
	if len(ids) == 0 {
		return true, nil
	}
	idList := make([]interface{}, len(ids))
	for i, id := range ids {
		idList[i] = id
	}
	_, err = orm.UnlinkWhere(ctx, model, [][]interface{}{{"id", "in", idList}})
	if err != nil {
		return nil, err
	}
	return true, nil
}
