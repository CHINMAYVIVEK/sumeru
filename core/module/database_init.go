package module

import (
	"context"
	"fmt"

	"sumeru/core/orm"
)

// RunFirstTimeInstallSync runs the DDL and sys.module seed steps shared with /setup/init
// (platform + base tables only, then registry rows for discovered addons). Call before
// InstallModuleByName(ctx, "base") and security bootstrap.
func RunFirstTimeInstallSync(ctx context.Context) error {
	if err := orm.SyncModelsInitialSetup(); err != nil {
		return fmt.Errorf("sync models initial: %w", err)
	}
	if err := orm.SyncRegistrySchemaInitialSetup(); err != nil {
		return fmt.Errorf("registry schema initial: %w", err)
	}
	if err := SyncSysModuleRegistry(ctx); err != nil {
		return fmt.Errorf("sync sys.module: %w", err)
	}
	return nil
}
