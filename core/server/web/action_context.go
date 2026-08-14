package web

import (
	"encoding/json"
	"strings"

	"sumeru/core/orm"
)

func actionViewIDFromContext(actionData map[string]interface{}) string {
	raw := strings.TrimSpace(orm.AsString(actionData["context"]))
	if raw == "" {
		return ""
	}
	var ctx map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(ctx["view_id"]))
}
