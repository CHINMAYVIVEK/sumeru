package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcCreate(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, newRPCError(CodeInvalidArgs, "create requires args[0] values object", map[string]interface{}{"method": "create"})
	}
	var vals map[string]interface{}
	if err := json.Unmarshal(arr[0], &vals); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[0] values: %v", err), map[string]interface{}{"method": "create"})
	}
	inst, ok := orm.Registry[model]
	if !ok || inst == nil {
		return nil, newRPCError(CodeModelNotFound, fmt.Sprintf("model %q not registered", model), map[string]interface{}{"model": model})
	}
	id, err := orm.Create(ctx, inst, vals)
	if err != nil {
		return nil, err
	}
	return id, nil
}
