package module

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"sumeru/addons/mail"
	"sumeru/core/metrics"
	"sumeru/core/orm"
)

// InstallModuleByName marks the module installed and loads its XML (and dependencies first).
func InstallModuleByName(context context.Context, moduleName string) error {
	if err := orm.CheckModelAccess(context, orm.SecurityUID(context), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(context, true)
	installMu.Lock()
	defer installMu.Unlock()
	return installModuleUnlocked(systemContext, moduleName)
}

func installModuleUnlocked(context context.Context, moduleName string) error {
	start := time.Now()
	defer func() {
		metrics.ObserveDuration("sumeru_module_install_duration_seconds", time.Since(start))
	}()
	if moduleName == "" {
		return fmt.Errorf("module name required")
	}
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	for _, dependencyName := range addon.Manifest.Depends {
		dependencyName = strings.TrimSpace(dependencyName)
		if dependencyName == "" || dependencyName == addon.Manifest.Name {
			continue
		}
		if _, has := DiscoveredAddons[dependencyName]; !has {
			continue
		}
		moduleRow, err := moduleRow(context, dependencyName)
		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("dependency %q is not registered", dependencyName)
			}
			return err
		}
		if moduleStateString(moduleRow) != "installed" {
			if err := installModuleUnlocked(context, dependencyName); err != nil {
				return fmt.Errorf("install dependency %q: %w", dependencyName, err)
			}
		}
	}

	if err := setModuleState(context, moduleName, "to_install", true); err != nil {
		return err
	}

	if err := orm.SyncRegistrySchemaForModule(moduleName); err != nil {
		_ = setModuleLastError(context, moduleName, err.Error())
		return FatalSync(moduleName, "schema sync", err)
	}

	if err := addon.SyncToDB(context); err != nil {
		_ = setModuleLastError(context, moduleName, err.Error())
		if IsFatalSync(err) {
			return err
		}
		// Recoverable warnings: still mark installed but keep last_error.
	} else {
		_ = setModuleLastError(context, moduleName, "")
	}

	if err := setModuleState(context, moduleName, "installed", true); err != nil {
		return err
	}
	orm.InvalidateRuleCache()
	mail.LogModuleEvent(context, moduleName, "Installed", "")
	return nil
}

func setModuleState(ctx context.Context, moduleName, state string, active bool) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET state = $1, active = $2 WHERE name = $3`,
		state, active, moduleName,
	)
	return err
}

func setModuleLastError(ctx context.Context, moduleName, msg string) error {
	_, err := orm.DB.ExecContext(ctx,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET last_error = $1 WHERE name = $2`,
		msg, moduleName,
	)
	return err
}

// UninstallModuleByName removes XML-linked metadata for the module and marks it uninstalled.
func UninstallModuleByName(context context.Context, moduleName string) error {
	if err := orm.CheckModelAccess(context, orm.SecurityUID(context), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(context, true)
	installMu.Lock()
	defer installMu.Unlock()

	if orm.IsPlatformModule(moduleName) {
		return fmt.Errorf("cannot uninstall platform module %q", moduleName)
	}
	if _, ok := DiscoveredAddons[moduleName]; !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	if dependency, err := installedModuleDependingOn(systemContext, moduleName); err != nil {
		return err
	} else if dependency != "" {
		return fmt.Errorf("module %q is required by installed module %q; uninstall that first", moduleName, dependency)
	}

	if _, err := orm.DB.ExecContext(systemContext,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET state = 'to_remove' WHERE name = $1`,
		moduleName,
	); err != nil {
		return err
	}

	if err := deleteModuleMetadata(systemContext, moduleName); err != nil {
		return err
	}

	if err := setModuleState(systemContext, moduleName, "uninstalled", true); err != nil {
		return err
	}
	mail.LogModuleEvent(systemContext, moduleName, "Uninstalled", "")
	return nil
}

func deleteModuleMetadata(context context.Context, moduleName string) error {
	modelNames := []string{"sys.menu", "sys.view", "sys.action.window", "sys.access", "sys.rule", "sys.approval.rule"}
	for _, modelName := range modelNames {
		tableName, err := orm.QuotedTableName(modelName)
		if err != nil {
			return fmt.Errorf("delete %s: %w", modelName, err)
		}
		deleteQuery := `DELETE FROM ` + tableName + ` WHERE id IN (SELECT core_id FROM ` + orm.MustQuotedTableName("sys.model.data") + ` WHERE module = $1 AND model = $2)`
		if _, err := orm.DB.ExecContext(context, deleteQuery, moduleName, modelName); err != nil {
			return fmt.Errorf("delete %s: %w", modelName, err)
		}
	}
	if _, err := orm.DB.ExecContext(context, `DELETE FROM `+orm.MustQuotedTableName("sys.model.data")+` WHERE module = $1`, moduleName); err != nil {
		return err
	}
	return nil
}

// SetModuleActive toggles visibility of menus for an installed module without removing data.
func SetModuleActive(context context.Context, moduleName string, active bool) error {
	if err := orm.CheckModelAccess(context, orm.SecurityUID(context), "sys.module", "write"); err != nil {
		return err
	}
	systemContext := orm.ContextWithBypass(context, true)
	installMu.Lock()
	defer installMu.Unlock()

	if moduleName == KernelModule && !active {
		return fmt.Errorf("cannot deactivate core module %q", KernelModule)
	}
	if _, ok := DiscoveredAddons[moduleName]; !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}

	moduleRow, err := moduleRow(systemContext, moduleName)
	if err != nil {
		return err
	}
	if moduleStateString(moduleRow) != "installed" {
		return fmt.Errorf("module %q is not installed; activate/install it first", moduleName)
	}

	if _, err := orm.DB.ExecContext(systemContext,
		`UPDATE `+orm.MustQuotedTableName("sys.module")+` SET active = $1 WHERE name = $2`,
		active, moduleName,
	); err != nil {
		return err
	}
	if active {
		mail.LogModuleEvent(systemContext, moduleName, "Activated", "")
	} else {
		mail.LogModuleEvent(systemContext, moduleName, "Deactivated", "")
	}
	return nil
}

// ListModules returns sys.module rows for the Apps UI (non-application modules included for completeness).
func ListModules(context context.Context) ([]map[string]interface{}, error) {
	moduleRows, err := orm.DB.QueryContext(context,
		`SELECT id, name, display_name, author, version, description, state, application, active FROM `+
			orm.MustQuotedTableName("sys.module")+` ORDER BY application DESC, name`,
	)
	if err != nil {
		return nil, err
	}
	defer moduleRows.Close()

	var moduleList []map[string]interface{}
	columnNames := []string{"id", "name", "display_name", "author", "version", "description", "state", "application", "active"}

	for moduleRows.Next() {
		var id int64
		var name, display, author, version, state string
		var desc sql.NullString
		var application, active bool
		if err := moduleRows.Scan(&id, &name, &display, &author, &version, &desc, &state, &application, &active); err != nil {
			return nil, err
		}
		moduleMap := make(map[string]interface{})
		moduleMap["id"] = id
		moduleMap[columnNames[1]] = name
		moduleMap[columnNames[2]] = display
		moduleMap[columnNames[3]] = author
		moduleMap[columnNames[4]] = version
		if desc.Valid {
			moduleMap[columnNames[5]] = desc.String
		} else {
			moduleMap[columnNames[5]] = ""
		}
		moduleMap[columnNames[6]] = state
		moduleMap[columnNames[7]] = application
		moduleMap[columnNames[8]] = active
		moduleList = append(moduleList, moduleMap)
	}
	return moduleList, moduleRows.Err()
}
