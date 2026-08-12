package api

import (
	"context"
	"encoding/json"
	"fmt"

	"sumeru/core/orm"
)

func rpcSearchRead(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, newRPCError(CodeInvalidArgs, "search_read requires args[0] domain and args[1] fields", map[string]interface{}{"method": "search_read"})
	}
	domain, err := parseDomainArg(arr[0])
	if err != nil {
		return nil, newRPCError(CodeInvalidArgs, err.Error(), map[string]interface{}{"method": "search_read", "hint": "args[0] domain"})
	}
	domain = orm.SubstituteDomainUID(domain, orm.UIDFromContext(ctx))
	var fields []string
	if err := json.Unmarshal(arr[1], &fields); err != nil {
		return nil, newRPCError(CodeInvalidArgs, fmt.Sprintf("args[1] fields: %v", err), map[string]interface{}{"method": "search_read"})
	}
	limit, offset := parseLimitOffset(kwargs)
	rows, err := searchLimitWithOffset(ctx, model, domain, limit, offset)
	if err != nil {
		return nil, err
	}
	return projectFields(rows, fields), nil
}
