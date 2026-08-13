package module

import (
	"context"
	"fmt"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// linkXMLRecord upserts sys.model.data for a synced XML record and logs failures.
func linkXMLRecord(ctx context.Context, moduleName, xmlID, model string, coreID int) error {
	if coreID <= 0 || xmlID == "" {
		return nil
	}
	_, err := orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
		"module":  moduleName,
		"name":    xmlID,
		"model":   model,
		"core_id": coreID,
	}, "name")
	if err != nil {
		syncWarn(ctx, "Failed to link XML record %s.%s (%s id=%d): %v", moduleName, xmlID, model, coreID, err)
	}
	return err
}

func syncWarn(ctx context.Context, format string, args ...interface{}) {
	if ctx == nil {
		ctx = context.Background()
	}
	msg := fmt.Sprintf(format, args...)
	applog.Warn(ctx, applog.Event{
		Message:   msg,
		Component: "module",
		Operation: "sync",
		Status:    "partial",
		Context:   map[string]interface{}{"detail": msg},
	})
}
