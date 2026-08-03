package orm

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

func skipAuditModel(model string) bool {
	switch model {
	case "sys.audit", "sys.session", "app.log", "core.user.log", "mail.message":
		return true
	default:
		return false
	}
}

// AppendAudit writes an immutable audit row (best-effort; never fails the caller).
func AppendAudit(ctx context.Context, action, model string, resID int64, before, after map[string]interface{}, detail string) {
	if DB == nil || strings.TrimSpace(action) == "" || skipAuditModel(model) {
		return
	}
	inst, ok := Registry["sys.audit"]
	if !ok {
		return
	}
	uid := SecurityUID(ctx)
	var beforeJSON, afterJSON string
	if before != nil {
		if b, err := json.Marshal(scrubAuditMap(before)); err == nil {
			beforeJSON = string(b)
		}
	}
	if after != nil {
		if b, err := json.Marshal(scrubAuditMap(after)); err == nil {
			afterJSON = string(b)
		}
	}
	vals := map[string]interface{}{
		"action":      action,
		"model":       strings.TrimSpace(model),
		"res_id":      resID,
		"before_json": beforeJSON,
		"after_json":  afterJSON,
		"detail":      detail,
		"create_date": time.Now().UTC().Format(time.RFC3339),
	}
	if uid > 0 {
		vals["user_id"] = uid
	}
	bypass := ContextWithBypass(ctx, true)
	_, _ = Create(bypass, inst, vals)
}

func scrubAuditMap(m map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(m))
	for k, v := range m {
		lk := strings.ToLower(k)
		if lk == "password" || lk == "key_hash" || strings.Contains(lk, "password") {
			out[k] = "***"
			continue
		}
		out[k] = v
	}
	return out
}

// LogAccessDeny records a permission denial in sys.audit.
func LogAccessDeny(ctx context.Context, model, op, detail string) {
	AppendAudit(ctx, "access_deny", model, 0, nil, nil, op+": "+detail)
}
