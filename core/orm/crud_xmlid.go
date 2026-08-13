package orm

import (
	"context"
	"fmt"
	"strings"
)

// ResolveXmlId returns the database ID for a given XML ID (module.name).
// The name segment may contain dots (e.g. base.action_core.company → module base, name action_core.company).
func ResolveXmlId(ctx context.Context, xmlID string) (int, string, error) {
	parts := strings.Split(xmlID, ".")
	module := ""
	name := xmlID
	if len(parts) >= 2 {
		module = parts[0]
		name = strings.Join(parts[1:], ".")
	}

	criteria := map[string]interface{}{"name": name}
	if module != "" {
		criteria["module"] = module
	}

	data, err := SearchOne(ctx, "sys.model.data", criteria)
	if err != nil {
		return 0, "", err
	}
	rid, ok := CoerceInt64(data["core_id"])
	if !ok {
		return 0, "", fmt.Errorf("invalid core_id in sys.model.data")
	}
	return int(rid), AsString(data["model"]), nil
}
