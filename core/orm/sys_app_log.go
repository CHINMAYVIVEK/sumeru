package orm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// AppLog stores application and module lifecycle audit lines (install, update, etc.).
// Internal user-to-user chatter uses mail.message only.
type AppLog struct {
	ID         int    `orm:"id"`
	ModuleName string `orm:"module_name"`
	Action     string `orm:"action"`
	Detail     string `orm:"detail"`
	Author     string `orm:"author"`
	CreateDate string `orm:"create_date"`
}

func (AppLog) ModelName() string { return "app.log" }
func (AppLog) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "module_name", Type: Char, Required: true, Index: true},
		{Name: "action", Type: Char, Required: true},
		{Name: "detail", Type: Text},
		{Name: "author", Type: Char},
		{Name: "create_date", Type: DateTime, Required: true},
	}
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
	_, err := Create(ctx, AppLog{}, vals)
	return err
}

func init() {
	RegisterModelWithModule(AppLog{}, "base")
}
