// Package api exposes Odoo-style JSON-RPC over HTTP for authenticated sessions.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"sumeru/core/orm"
)

// rpcRequest is the JSON body for POST /api/rpc.
// If model is empty and params is set, params is unmarshalled again (Odoo-style wrapper).
type rpcRequest struct {
	Model  string          `json:"model"`
	Method string          `json:"method"`
	Args   json.RawMessage `json:"args"`
	Kwargs json.RawMessage `json:"kwargs"`
	Params json.RawMessage `json:"params"`
}

// DispatchRPC executes a model method call with the same security context as the web UI (uid in ctx).
// Supported methods: search, search_read, read, create, write, unlink.
func DispatchRPC(ctx context.Context, body []byte) (interface{}, error) {
	var in rpcRequest
	if err := json.Unmarshal(body, &in); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}
	if strings.TrimSpace(in.Model) == "" && len(in.Params) > 0 {
		if err := json.Unmarshal(in.Params, &in); err != nil {
			return nil, fmt.Errorf("invalid params: %w", err)
		}
	}
	model := strings.TrimSpace(in.Model)
	method := strings.TrimSpace(in.Method)
	if model == "" {
		return nil, fmt.Errorf("model is required")
	}
	if method == "" {
		return nil, fmt.Errorf("method is required")
	}
	if _, ok := orm.Registry[model]; !ok {
		return nil, fmt.Errorf("unknown model %q", model)
	}

	switch method {
	case "search":
		return rpcSearch(ctx, model, in.Args, in.Kwargs)
	case "search_read":
		return rpcSearchRead(ctx, model, in.Args, in.Kwargs)
	case "read":
		return rpcRead(ctx, model, in.Args)
	case "create":
		return rpcCreate(ctx, model, in.Args)
	case "write":
		return rpcWrite(ctx, model, in.Args)
	case "unlink":
		return rpcUnlink(ctx, model, in.Args)
	default:
		return nil, fmt.Errorf("unsupported method %q", method)
	}
}
