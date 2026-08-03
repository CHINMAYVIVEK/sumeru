package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/sdk/platformmsg"
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
		fmt.Printf("Warning: sys.action.window record %s (module %s): core_model is required\n", xmlRecord.ID, moduleName)
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

func (addon *Addon) SyncToDB(ctx context.Context) error {
	moduleName := addon.Manifest.Name

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
			return err
		}
	}

	if err := addon.syncCSVModelAccess(ctx); err != nil {
		fmt.Printf("Warning: Failed to load CSV ACLs for %s: %v\n", moduleName, err)
	}
	var inheritQueue []parser.Record

	for _, xmlFile := range addon.Manifest.Data {
		xmlPath := filepath.Join(addon.Path, xmlFile)
		if _, err := os.Stat(xmlPath); err != nil {
			fmt.Printf(platformmsg.FmtDataFileMissing, xmlFile, moduleName)
			continue
		}

		parsedViewData, err := parser.ParseViewList(xmlPath)
		if err == nil && (len(parsedViewData.Records) > 0 || len(parsedViewData.Views) > 0) {
			for _, xmlRecord := range parsedViewData.Records {
				if xmlRecord.Model == "sys.action.window" {
					upsertSysActionWindowFromRecord(ctx, moduleName, xmlRecord)
				}
				if xmlRecord.Model == "sys.view" {
					if strings.TrimSpace(parser.RecordFieldMap(xmlRecord)["inherit_id"]) != "" {
						inheritQueue = append(inheritQueue, xmlRecord)
					} else {
						upsertSysViewFromRecord(ctx, moduleName, xmlRecord)
					}
				}
				syncGenericRegistryRecord(ctx, moduleName, xmlRecord)
			}

			for _, viewDef := range parsedViewData.Views {
				upsertInlineViewDef(ctx, moduleName, &viewDef)
			}
			continue
		}

		menuList, err := parser.ParseMenuList(xmlPath)
		if err == nil {
			if len(menuList.MenuItems) > 0 {
				syncMenusFromItems(ctx, moduleName, menuList.MenuItems)
			}
			for _, xmlRecord := range menuList.Records {
				if xmlRecord.Model == "sys.action.window" {
					upsertSysActionWindowFromRecord(ctx, moduleName, xmlRecord)
				}
				if xmlRecord.Model == "sys.view" {
					if strings.TrimSpace(parser.RecordFieldMap(xmlRecord)["inherit_id"]) != "" {
						inheritQueue = append(inheritQueue, xmlRecord)
					} else {
						upsertSysViewFromRecord(ctx, moduleName, xmlRecord)
					}
				}
				syncGenericRegistryRecord(ctx, moduleName, xmlRecord)
			}
		}
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			fmt.Printf(platformmsg.FmtViewInheritWarning, moduleName, xmlRecord.ID, err)
		}
	}

	return nil
}
