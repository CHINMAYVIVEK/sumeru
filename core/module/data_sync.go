package module

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/base/platformmsg"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

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
					}
				}
				syncGenericRegistryRecord(ctx, moduleName, xmlRecord)
			}

			for _, viewDef := range parsedViewData.Views {
				upsertInlineViewDef(ctx, moduleName, &viewDef)
			}
			continue
		}

		menus, err := parser.ParseMenus(xmlPath)
		if err == nil && len(menus) > 0 {
			syncMenusFromItems(ctx, moduleName, menus)
		}
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(ctx, moduleName, xmlRecord); err != nil {
			fmt.Printf(platformmsg.FmtViewInheritWarning, moduleName, xmlRecord.ID, err)
		}
	}

	return nil
}
