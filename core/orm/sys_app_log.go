package orm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sumeru/core/modelmeta"
)

type AppLog struct {
	modelmeta.ModelMeta `sumeru:"model=app.log"`

	ModuleName modelmeta.String `sumeru:"required,index,column=module_name"`
	Action     modelmeta.String `sumeru:"required"`
	Detail     modelmeta.Text
	Author     modelmeta.String
	CreateDate modelmeta.DateTime `sumeru:"required,column=create_date"`
}

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
	_, err := Create(ctx, RegistryModel("app.log"), vals)
	return err
}
