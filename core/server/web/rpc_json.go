package web

import (
	"encoding/json"
	"io"
	"net/http"

	"sumeru/core/applog"
	"sumeru/core/orm"
	"sumeru/core/server/api"
)

const maxRPCBody = 4 << 20

// APIHealthHandler returns {"ok":true} for probes (no auth).
func APIHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// RPCJSONHandler is model RPC: POST JSON {"model","method","args","kwargs"} with session or API key auth.
func RPCJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	uid := AuthenticatedUserID(r)
	if uid <= 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRPCBody))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	ctx := orm.ContextWithUID(r.Context(), uid)
	result, err := api.DispatchRPC(ctx, body)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		applog.L(ctx).Warnw("api.rpc", "err", err)
		writeRPCEnvelope(w, nil, err)
		return
	}
	writeRPCEnvelope(w, result, nil)
}

func writeRPCEnvelope(w http.ResponseWriter, result interface{}, err error) {
	enc := json.NewEncoder(w)
	if err != nil {
		_ = enc.Encode(map[string]interface{}{
			"result": nil,
			"error": map[string]string{
				"message": err.Error(),
				"code":    "RPC_ERROR",
			},
		})
		return
	}
	_ = enc.Encode(map[string]interface{}{"result": result, "error": nil})
}
