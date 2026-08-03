package module

import (
	"context"
	"database/sql"
	"fmt"

	"sumeru/core/orm"
)

func countSysModules(context context.Context) (int, error) {
	countRow := orm.DB.QueryRowContext(context, "SELECT COUNT(*) FROM "+orm.GetTableName("sys.module"))
	var moduleCount int
	if err := countRow.Scan(&moduleCount); err != nil {
		return 0, err
	}
	return moduleCount, nil
}

// SyncSysModuleRegistry upserts sys.module rows for every discovered addon (same as normal startup).
// Call during /setup after core tables exist so InstallModuleByName can update module state.
func SyncSysModuleRegistry(context context.Context) error {
	return syncSysModuleRows(context, DiscoveredAddons)
}

// syncSysModuleRows upserts registry metadata. New DB: only the base module row is inserted as installed;
// every other discovered addon is uninstalled until explicitly installed from Apps or CLI.
func syncSysModuleRows(context context.Context, discovered map[string]*Addon) error {
	totalModules, err := countSysModules(context)
	if err != nil {
		return err
	}
	isBootstrap := totalModules == 0

	for _, addon := range discovered {
		_, err := orm.SearchOne(context, "sys.module", map[string]interface{}{"name": addon.Manifest.Name})
		if err == sql.ErrNoRows {
			state := "uninstalled"
			if isBootstrap {
				// Platform spine (base, mail, …) auto-installs; other addons start uninstalled.
				if IsPlatformAutoInstall(addon.Manifest.Name) {
					state = "installed"
				}
			}
			_, err = orm.Create(context, orm.SysModule{}, map[string]interface{}{
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
		_, err = orm.DB.ExecContext(context,
			`UPDATE `+orm.GetTableName("sys.module")+
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

func moduleRow(context context.Context, moduleName string) (map[string]interface{}, error) {
	return orm.SearchOne(context, "sys.module", map[string]interface{}{"name": moduleName})
}

func shouldSyncData(context context.Context, moduleName string) bool {
	moduleRow, err := moduleRow(context, moduleName)
	if err != nil {
		return false
	}
	state, _ := moduleRow["state"].(string)
	active, _ := moduleRow["active"].(bool)
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
