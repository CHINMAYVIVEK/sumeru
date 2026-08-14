package module

import (
	"context"
	"database/sql"
	"fmt"

	"sumeru/core/orm"
)

func countSysModules(ctx context.Context) (int, error) {
	countRow := orm.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+orm.MustQuotedTableName("sys.module"))
	var moduleCount int
	if err := countRow.Scan(&moduleCount); err != nil {
		return 0, err
	}
	return moduleCount, nil
}

// SyncSysModuleRegistry upserts sys.module rows for every discovered addon (same as normal startup).
// Call during /setup after core tables exist so InstallModuleByName can update module state.
func SyncSysModuleRegistry(ctx context.Context) error {
	return syncSysModuleRows(ctx, DiscoveredAddons)
}

// syncSysModuleRows upserts registry metadata. New DB: only the base module row is inserted as installed;
// every other discovered addon is uninstalled until explicitly installed from Apps or CLI.
func syncSysModuleRows(ctx context.Context, discovered map[string]*Addon) error {
	totalModules, err := countSysModules(ctx)
	if err != nil {
		return err
	}
	isBootstrap := totalModules == 0

	for _, addon := range discovered {
		_, err := orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": addon.Manifest.Name})
		if err == sql.ErrNoRows {
			state := "uninstalled"
			if isBootstrap {
				if orm.IsPlatformModule(addon.Manifest.Name) {
					state = "installed"
				}
			}
			_, err = orm.Create(ctx, orm.SysModule{}, map[string]interface{}{
				"name":         addon.Manifest.Name,
				"display_name": irModuleDisplayName(addon),
				"author":       addon.Manifest.Author,
				"version":      addon.Manifest.Version,
				"description":  addon.Manifest.Description,
				"state":        state,
				"application":  addon.Manifest.IsApplication(),
				"active":       true,
			})
			if err != nil {
				return fmt.Errorf("create sys.module %s: %w", addon.Manifest.Name, err)
			}
			continue
		}
		if err != nil {
			return err
		}
		_, err = orm.DB.ExecContext(ctx,
			`UPDATE `+orm.MustQuotedTableName("sys.module")+
				` SET display_name = $1, author = $2, version = $3, description = $4, application = $5 WHERE name = $6`,
			irModuleDisplayName(addon),
			addon.Manifest.Author,
			addon.Manifest.Version,
			addon.Manifest.Description,
			addon.Manifest.IsApplication(),
			addon.Manifest.Name,
		)
		if err != nil {
			return fmt.Errorf("update sys.module %s: %w", addon.Manifest.Name, err)
		}
	}
	return nil
}

func moduleRow(ctx context.Context, moduleName string) (map[string]interface{}, error) {
	return orm.SearchOne(ctx, "sys.module", map[string]interface{}{"name": moduleName})
}

func shouldSyncData(ctx context.Context, moduleName string) bool {
	row, err := moduleRow(ctx, moduleName)
	if err != nil {
		return false
	}
	state, _ := row["state"].(string)
	active, _ := row["active"].(bool)
	if !active {
		return false
	}
	return state == "installed"
}

func moduleStateString(moduleRow map[string]interface{}) string {
	if moduleRow == nil {
		return ""
	}
	switch value := moduleRow["state"].(type) {
	case string:
		return value
	case []byte:
		return string(value)
	default:
		return fmt.Sprint(value)
	}
}
