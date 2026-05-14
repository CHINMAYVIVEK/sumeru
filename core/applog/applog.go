package applog

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/orm"
)

// Log appends one application or module lifecycle row to app.log.
func Log(ctx context.Context, moduleName, action, detail string) error {
	if orm.DB == nil {
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
	_, err := orm.Create(ctx, orm.AppLog{}, vals)
	return err
}
