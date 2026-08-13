package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/metrics"
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
	start := time.Now()
	metrics.Inc("sumeru_rpc_requests_total")
	defer func() {
		metrics.ObserveDuration("sumeru_rpc_duration_seconds", time.Since(start))
	}()

	if r.Method != http.MethodPost {
		api.WriteResponse(w, http.StatusMethodNotAllowed, api.Fail(api.CodeMethodNotAllowed, "Method not allowed", nil))
		return
	}
	uid := AuthenticatedUserID(r)
	if uid <= 0 {
		api.WriteResponse(w, http.StatusUnauthorized, api.Fail(api.CodeUnauthorized, "Unauthorized", nil))
		return
	}
	ct := strings.TrimSpace(strings.ToLower(r.Header.Get("Content-Type")))
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		api.WriteResponse(w, http.StatusUnsupportedMediaType, api.Fail(api.CodeUnsupportedMediaType, "Content-Type must be application/json", nil))
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxRPCBody+1))
	if err != nil {
		api.WriteResponse(w, http.StatusBadRequest, api.Fail(api.CodeInvalidBody, "Could not read request body", nil))
		return
	}
	if len(body) > maxRPCBody {
		api.WriteResponse(w, http.StatusRequestEntityTooLarge, api.Fail(api.CodePayloadTooLarge, "Request body too large", nil))
		return
	}
	ctx := orm.ContextWithUID(r.Context(), uid)
	if cid := orm.CompanyIDFromContext(r.Context()); cid > 0 {
		ctx = orm.ContextWithCompanyID(ctx, cid)
	} else if uid > 0 {
		ctx = orm.ContextWithCompanyID(ctx, orm.ActiveCompanyIDForUser(ctx, uid))
	}
	resp, status := api.Dispatch(ctx, body)

	modelName, methodName := rpcModelMethod(body)
	ev := applog.Event{
		Component: "rpc",
		Operation: methodName,
		Duration:  time.Since(start),
		Context: map[string]interface{}{
			"resource":    modelName,
			"method":      methodName,
			"status_code": status,
		},
	}
	if !resp.OK && resp.Error != nil {
		ev.Message = "RPC call failed"
		ev.Status = "failure"
		ev.Context["error_code"] = resp.Error.Code
		ev.Context["error"] = resp.Error.Message
		applog.Error(ctx, ev)
	} else {
		ev.Message = "RPC call completed"
		ev.Status = "success"
		applog.Info(ctx, ev)
	}
	api.WriteResponse(w, status, resp)
}

func rpcModelMethod(body []byte) (model, method string) {
	var req struct {
		Model  string `json:"model"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		return "", ""
	}
	return req.Model, req.Method
}
