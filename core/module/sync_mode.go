package module

import (
	"context"

	"sumeru/core/orm"
)

type syncModeKey struct{}

// ContextWithSyncMode attaches install/update reload mode to ctx for data sync.
func ContextWithSyncMode(ctx context.Context, mode moduleReloadMode) context.Context {
	return context.WithValue(ctx, syncModeKey{}, mode)
}

// SyncModeFromContext returns moduleReloadInstall when unset.
func SyncModeFromContext(ctx context.Context) moduleReloadMode {
	if v, ok := ctx.Value(syncModeKey{}).(moduleReloadMode); ok {
		return v
	}
	return moduleReloadInstall
}

type dataFileOpts struct {
	noUpdate bool
}

func (o dataFileOpts) skipExistingOnUpdate(ctx context.Context, moduleName, xmlID string) bool {
	if SyncModeFromContext(ctx) != moduleReloadUpdate || !o.noUpdate || xmlID == "" {
		return false
	}
	existingID, _, err := orm.ResolveXmlId(ctx, moduleName+"."+xmlID)
	return err == nil && existingID > 0
}
