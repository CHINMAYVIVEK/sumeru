package api

import (
	"context"
	"encoding/json"

	"sumeru/core/orm"
)

func rpcSearch(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	var domain [][]interface{}
	if len(arr) >= 1 {
		domain, err = parseDomainArg(arr[0])
		if err != nil {
			return nil, newRPCError(CodeInvalidArgs, err.Error(), map[string]interface{}{"method": "search", "hint": "args[0] domain"})
		}
	}
	domain = orm.SubstituteDomainUID(domain, orm.UIDFromContext(ctx))
	limit, offset := parseLimitOffset(kwargs)
	return searchLimitWithOffset(ctx, model, domain, limit, offset)
}
