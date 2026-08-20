package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func processXMLRecords(ctx context.Context, moduleName string, records []parser.Record, inheritQueue *[]parser.Record, opts dataFileOpts) {
	for _, xmlRecord := range records {
		if opts.skipExistingOnUpdate(ctx, moduleName, xmlRecord.ID) {
			continue
		}
		if xmlRecord.Model == "sys.action.window" {
			upsertSysActionWindowFromRecord(ctx, moduleName, xmlRecord)
		}
		if xmlRecord.Model == "sys.view" {
			if strings.TrimSpace(parser.RecordFieldMap(xmlRecord)["inherit_id"]) != "" {
				*inheritQueue = append(*inheritQueue, xmlRecord)
			} else {
				upsertSysViewFromRecord(ctx, moduleName, xmlRecord)
			}
		}
		syncGenericRegistryRecord(ctx, moduleName, xmlRecord)
	}
}

func RecordsFromActions(actions []parser.Action) []parser.Record {
	if len(actions) == 0 {
		return nil
	}
	out := make([]parser.Record, 0, len(actions))
	for _, a := range actions {
		out = append(out, a.ToRecord())
	}
	return out
}

// loadManifestDataFile parses one manifest XML path (view or menu layout) and syncs its content.
// Menu items are appended to menuCollector for a deferred sync pass after all manifest files load.
func loadManifestDataFile(ctx context.Context, moduleName, xmlPath, relFile string, inheritQueue *[]parser.Record, menuCollector *[]parser.MenuItem) []error {
	parsedViewData, viewErr := parser.ParseViewList(xmlPath)
	if viewErr == nil {
		opts := dataFileOpts{noUpdate: parsedViewData.NoUpdate}
		records := append([]parser.Record(nil), parsedViewData.Records...)
		records = append(records, RecordsFromActions(parsedViewData.Actions)...)
		if len(records) > 0 || len(parsedViewData.Views) > 0 || len(parsedViewData.MenuItems) > 0 {
			processXMLRecords(ctx, moduleName, records, inheritQueue, opts)
			for i := range parsedViewData.Views {
				upsertInlineViewDef(ctx, moduleName, &parsedViewData.Views[i])
			}
			if len(parsedViewData.MenuItems) > 0 {
				*menuCollector = append(*menuCollector, parsedViewData.MenuItems...)
			}
			return nil
		}
	} else if !AllowMenuParseFallback(viewErr) {
		return []error{RecoverableSync(moduleName, "ParseViewList "+relFile, viewErr)}
	}

	menuList, menuErr := parser.ParseMenuList(xmlPath)
	if menuErr == nil {
		opts := dataFileOpts{noUpdate: menuList.NoUpdate}
		records := append([]parser.Record(nil), menuList.Records...)
		records = append(records, RecordsFromActions(menuList.Actions)...)
		if len(menuList.MenuItems) > 0 || len(records) > 0 {
			if len(menuList.MenuItems) > 0 {
				*menuCollector = append(*menuCollector, menuList.MenuItems...)
			}
			processXMLRecords(ctx, moduleName, records, inheritQueue, opts)
			return nil
		}
		return nil
	}

	if viewErr != nil {
		return []error{RecoverableSync(moduleName, "parse "+relFile,
			fmt.Errorf("ParseViewList: %v; ParseMenuList: %v", viewErr, menuErr))}
	}
	return []error{RecoverableSync(moduleName, "ParseMenuList "+relFile, menuErr)}
}

// CollectMenuItemsFromManifestFile parses menu items from one manifest XML path (for tests and tooling).
func CollectMenuItemsFromManifestFile(xmlPath string) ([]parser.MenuItem, error) {
	parsedViewData, viewErr := parser.ParseViewList(xmlPath)
	if viewErr == nil && len(parsedViewData.MenuItems) > 0 {
		return append([]parser.MenuItem(nil), parsedViewData.MenuItems...), nil
	}
	if viewErr != nil && !AllowMenuParseFallback(viewErr) {
		return nil, viewErr
	}
	menuList, menuErr := parser.ParseMenuList(xmlPath)
	if menuErr != nil {
		if viewErr != nil {
			return nil, fmt.Errorf("ParseViewList: %v; ParseMenuList: %v", viewErr, menuErr)
		}
		return nil, menuErr
	}
	return append([]parser.MenuItem(nil), menuList.MenuItems...), nil
}

func AllowMenuParseFallback(viewErr error) bool {
	if viewErr == nil {
		return true
	}
	msg := viewErr.Error()
	// Invalid module root will fail menu parse the same way — do not double-warn.
	if strings.Contains(msg, "module XML root must be") {
		return false
	}
	return true
}

func (addon *Addon) SyncToDB(ctx context.Context) error {
	moduleName := addon.Manifest.Name
	var errs []error

	for _, registeredModel := range orm.Registry {
		if strings.TrimSpace(orm.DeclaringModule(registeredModel.ModelName())) != moduleName {
			continue
		}
		_, err := orm.Upsert(ctx, orm.RegistryModel("sys.model"), map[string]interface{}{
			"name":   registeredModel.ModelName(),
			"model":  registeredModel.ModelName(),
			"module": moduleName,
		}, "name")
		if err != nil {
			errs = append(errs, FatalSync(moduleName, "sys.model upsert "+registeredModel.ModelName(), err))
		}
	}

	if err := addon.syncCSVModelAccess(ctx); err != nil {
		errs = append(errs, FatalSync(moduleName, "CSV ACL load", err))
	} else {
		orm.InvalidateRuleCache()
	}
	var inheritQueue []parser.Record
	var deferredMenus []parser.MenuItem

	for _, xmlFile := range addon.Manifest.Data {
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(xmlFile)), ".csv") {
			continue // ACL CSV is loaded by syncCSVModelAccess above
		}
		xmlPath := filepath.Join(addon.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "data file missing "+xmlFile, err))
			continue
		}

		if fileErrs := loadManifestDataFile(ctx, moduleName, xmlPath, xmlFile, &inheritQueue, &deferredMenus); len(fileErrs) > 0 {
			errs = append(errs, fileErrs...)
		}
	}

	if len(deferredMenus) > 0 {
		syncMenusFromItems(ctx, moduleName, deferredMenus)
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "view inherit "+xmlRecord.ID, err))
		}
	}

	return aggregateErrors(moduleName, errs)
}
