package module

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// resolveXMLIDInModule resolves module.external_id: uses full ref if it contains a dot,
// else current module's xml id, then a global name-only lookup.
func resolveXMLIDInModule(ctx context.Context, moduleName, ref string) (int, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return 0, nil
	}
	if strings.Contains(ref, ".") {
		id, _, err := orm.ResolveXmlId(ctx, ref)
		return id, err
	}
	if moduleName != "" {
		if id, _, err := orm.ResolveXmlId(ctx, moduleName+"."+ref); id != 0 {
			return id, nil
		} else if err != nil {
			return 0, err
		}
	}
	id, _, err := orm.ResolveXmlId(ctx, ref)
	return id, err
}

// linkXMLRecord upserts sys.model.data for a synced XML record and logs failures.
func linkXMLRecord(ctx context.Context, moduleName, xmlID, model string, coreID int) error {
	if coreID <= 0 || xmlID == "" {
		return nil
	}
	_, err := orm.Upsert(ctx, orm.RegistryModel("sys.model.data"), map[string]interface{}{
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
