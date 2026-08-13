package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func upsertSysActionWindowFromRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	fieldMap := parser.RecordFieldMap(xmlRecord)
	recordValues := map[string]interface{}{}
	for key, val := range fieldMap {
		recordValues[key] = val
	}
	if cm := strings.TrimSpace(orm.AsString(recordValues["core_model"])); cm == "" {
		applog.L(context.Background()).Warn("module_sync", "msg", fmt.Sprintf("Warning: sys.action.window record %s (module %s): core_model is required", xmlRecord.ID, moduleName))
		return
	}
	if _, ok := recordValues["name"]; !ok || recordValues["name"] == "" {
		recordValues["name"] = xmlRecord.ID
	}
	id, err := orm.Upsert(ctx, orm.SysActionWindow{}, recordValues, "name")
	if err == nil {
		_ = linkXMLRecord(ctx, moduleName, xmlRecord.ID, "sys.action.window", id)
	}
}

func processXMLRecords(ctx context.Context, moduleName string, records []parser.Record, inheritQueue *[]parser.Record) {
	for _, xmlRecord := range records {
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
func loadManifestDataFile(ctx context.Context, moduleName, xmlPath, relFile string, inheritQueue *[]parser.Record) []error {
	parsedViewData, viewErr := parser.ParseViewList(xmlPath)
	if viewErr == nil {
		records := append([]parser.Record(nil), parsedViewData.Records...)
		records = append(records, RecordsFromActions(parsedViewData.Actions)...)
		if len(records) > 0 || len(parsedViewData.Views) > 0 || len(parsedViewData.MenuItems) > 0 {
			processXMLRecords(ctx, moduleName, records, inheritQueue)
			for i := range parsedViewData.Views {
				upsertInlineViewDef(ctx, moduleName, &parsedViewData.Views[i])
			}
			if len(parsedViewData.MenuItems) > 0 {
				syncMenusFromItems(ctx, moduleName, parsedViewData.MenuItems)
			}
			return nil
		}
	} else if !AllowMenuParseFallback(viewErr) {
		return []error{RecoverableSync(moduleName, "ParseViewList "+relFile, viewErr)}
	}

	menuList, menuErr := parser.ParseMenuList(xmlPath)
	if menuErr == nil {
		records := append([]parser.Record(nil), menuList.Records...)
		records = append(records, RecordsFromActions(menuList.Actions)...)
		if len(menuList.MenuItems) > 0 || len(records) > 0 {
			if len(menuList.MenuItems) > 0 {
				syncMenusFromItems(ctx, moduleName, menuList.MenuItems)
			}
			processXMLRecords(ctx, moduleName, records, inheritQueue)
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
		_, err := orm.Upsert(ctx, orm.SysModel{}, map[string]interface{}{
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

	for _, xmlFile := range addon.Manifest.Data {
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(xmlFile)), ".csv") {
			continue // ACL CSV is loaded by syncCSVModelAccess above
		}
		xmlPath := filepath.Join(addon.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "data file missing "+xmlFile, err))
			continue
		}

		if fileErrs := loadManifestDataFile(ctx, moduleName, xmlPath, xmlFile, &inheritQueue); len(fileErrs) > 0 {
			errs = append(errs, fileErrs...)
		}
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "view inherit "+xmlRecord.ID, err))
		}
	}

	return aggregateErrors(moduleName, errs)
}
