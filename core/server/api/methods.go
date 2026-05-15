package api

import (
	"context"
	"encoding/json"
	"fmt"

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
			return nil, err
		}
	}
	domain = orm.SubstituteDomainUID(domain, orm.UIDFromContext(ctx))
	limit, _ := parseLimitOffset(kwargs)
	return orm.SearchLimit(ctx, model, domain, limit)
}

func rpcSearchRead(ctx context.Context, model string, args, kwargs json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, fmt.Errorf("search_read requires args[0] domain and args[1] fields")
	}
	domain, err := parseDomainArg(arr[0])
	if err != nil {
		return nil, err
	}
	domain = orm.SubstituteDomainUID(domain, orm.UIDFromContext(ctx))
	var fields []string
	if err := json.Unmarshal(arr[1], &fields); err != nil {
		return nil, fmt.Errorf("args[1] fields: %w", err)
	}
	limit, _ := parseLimitOffset(kwargs)
	rows, err := orm.SearchLimit(ctx, model, domain, limit)
	if err != nil {
		return nil, err
	}
	return projectFields(rows, fields), nil
}

func projectFields(rows []map[string]interface{}, fields []string) []map[string]interface{} {
	if len(fields) == 0 {
		return rows
	}
	set := map[string]struct{}{}
	for _, f := range fields {
		set[f] = struct{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		m := make(map[string]interface{}, len(fields))
		for k, v := range r {
			if _, ok := set[k]; ok {
				m[k] = v
			}
		}
		out = append(out, m)
	}
	return out
}

func rpcRead(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, fmt.Errorf("read requires args[0] ids")
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, fmt.Errorf("ids: %w", err)
	}
	var fields []string
	if len(arr) >= 2 && len(arr[1]) > 0 && string(arr[1]) != "null" {
		if err := json.Unmarshal(arr[1], &fields); err != nil {
			return nil, err
		}
	}
	var out []map[string]interface{}
	for _, id := range ids {
		rec, err := orm.SearchOne(ctx, model, map[string]interface{}{"id": id})
		if err != nil {
			continue
		}
		if len(fields) > 0 {
			out = append(out, projectFields([]map[string]interface{}{rec}, fields)[0])
		} else {
			out = append(out, rec)
		}
	}
	return out, nil
}

func rpcCreate(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, fmt.Errorf("create requires args[0] values object")
	}
	var vals map[string]interface{}
	if err := json.Unmarshal(arr[0], &vals); err != nil {
		return nil, err
	}
	inst, ok := orm.Registry[model]
	if !ok || inst == nil {
		return nil, fmt.Errorf("model %q not registered", model)
	}
	id, err := orm.Create(ctx, inst, vals)
	if err != nil {
		return nil, err
	}
	return id, nil
}

func rpcWrite(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 2 {
		return nil, fmt.Errorf("write requires args[0] ids and args[1] values")
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, err
	}
	var vals map[string]interface{}
	if err := json.Unmarshal(arr[1], &vals); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if err := orm.UpdateRecordByID(ctx, model, id, vals); err != nil {
			return nil, err
		}
	}
	return true, nil
}

func rpcUnlink(ctx context.Context, model string, args json.RawMessage) (interface{}, error) {
	arr, err := parseArgsArray(args)
	if err != nil {
		return nil, err
	}
	if len(arr) < 1 {
		return nil, fmt.Errorf("unlink requires args[0] ids")
	}
	var ids []int
	if err := json.Unmarshal(arr[0], &ids); err != nil {
		return nil, err
	}
	for _, id := range ids {
		if err := orm.Unlink(ctx, model, id); err != nil {
			return nil, err
		}
	}
	return true, nil
}
