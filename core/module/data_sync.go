package module

import (
	"context"
	"sumeru/core/applog"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/engine/parser"
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
		_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
			"module":  moduleName,
			"name":    xmlRecord.ID,
			"model":   "sys.action.window",
			"core_id": id,
		}, "name")
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
		xmlPath := filepath.Join(addon.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "data file missing "+xmlFile, err))
			continue
		}

		parsedViewData, err := parser.ParseViewList(xmlPath)
		if err == nil {
			hasContent := len(parsedViewData.Records) > 0 || len(parsedViewData.Views) > 0 ||
				len(parsedViewData.MenuItems) > 0 || len(parsedViewData.Actions) > 0
			if hasContent {
				processXMLRecords(ctx, moduleName, parsedViewData.Records, &inheritQueue)
				for _, viewDef := range parsedViewData.Views {
					upsertInlineViewDef(ctx, moduleName, &viewDef)
				}
				if len(parsedViewData.MenuItems) > 0 {
					syncMenusFromItems(ctx, moduleName, parsedViewData.MenuItems)
				}
				continue
			}
		} else {
			errs = append(errs, RecoverableSync(moduleName, "ParseViewList "+xmlFile, err))
		}

		menuList, err := parser.ParseMenuList(xmlPath)
		if err == nil {
			if len(menuList.MenuItems) > 0 {
				syncMenusFromItems(ctx, moduleName, menuList.MenuItems)
			}
			processXMLRecords(ctx, moduleName, menuList.Records, &inheritQueue)
		} else if err != nil {
			errs = append(errs, RecoverableSync(moduleName, "ParseMenuList "+xmlFile, err))
		}
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			errs = append(errs, RecoverableSync(moduleName, "view inherit "+xmlRecord.ID, err))
		}
	}

	return aggregateErrors(moduleName, errs)
}
