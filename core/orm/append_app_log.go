package orm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AppendAppLog inserts one row into app.log (module lifecycle / audit).
func AppendAppLog(ctx context.Context, moduleName, action, detail string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}
	moduleName = strings.TrimSpace(moduleName)
	action = strings.TrimSpace(action)
	if moduleName == "" || action == "" {
		return fmt.Errorf("module name and action are required")
	}
	detail = strings.TrimSpace(detail)
	vals := map[string]interface{}{
		"module_name": moduleName,
		"action":      action,
		"detail":      detail,
		"author":      "System",
		"create_date": time.Now().UTC(),
	}
	_, err := Create(ctx, AppLog{}, vals)
	return err
}
