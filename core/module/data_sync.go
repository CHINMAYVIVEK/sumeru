package module

import (
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sumeru/core/base/platformmsg"
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

func (addon *Addon) SyncToDB(context context.Context) error {
	moduleName := addon.Manifest.Name

	for _, registeredModel := range orm.Registry {
		if strings.TrimSpace(orm.DeclaringModule(registeredModel.ModelName())) != moduleName {
			continue
		}
		_, err := orm.Upsert(context, orm.SysModel{}, map[string]interface{}{
			"name":   registeredModel.ModelName(),
			"model":  registeredModel.ModelName(),
			"module": moduleName,
		}, "name")
		if err != nil {
			return err
		}
	}

	// Load sys.access.csv if it exists
	if err := addon.syncCSVModelAccess(context); err != nil {
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
					fieldMap := parser.RecordFieldMap(xmlRecord)
					recordValues := map[string]interface{}{}
					for key, val := range fieldMap {
						recordValues[key] = val
					}
					if core := strings.TrimSpace(orm.AsString(recordValues["core_model"])); core == "" {
						if legacy := strings.TrimSpace(orm.AsString(recordValues["res_model"])); legacy != "" {
							recordValues["core_model"] = legacy
						}
					}
					delete(recordValues, "res_model")
					if _, ok := recordValues["name"]; !ok || recordValues["name"] == "" {
						recordValues["name"] = xmlRecord.ID
					}
					id, err := orm.Upsert(context, orm.SysActionWindow{}, recordValues, "name")
					if err == nil {
						orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
							"module":  moduleName,
							"name":    xmlRecord.ID,
							"model":   "sys.action.window",
							"core_id": id,
						}, "name")
					}
				}
				if xmlRecord.Model == "sys.view" {
					if strings.TrimSpace(parser.RecordFieldMap(xmlRecord)["inherit_id"]) != "" {
						inheritQueue = append(inheritQueue, xmlRecord)
					}
				}
				syncGenericRegistryRecord(context, moduleName, xmlRecord)
			}

			for _, viewDef := range parsedViewData.Views {
				viewArchitecture := viewArchXML(&viewDef)

				id, err := orm.Upsert(context, orm.SysView{}, map[string]interface{}{
					"name":  viewDef.Model + "." + viewDef.Type,
					"model": viewDef.Model,
					"type":  viewDef.Type,
					"arch":  viewArchitecture,
				}, "name")
				if err == nil {
					orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
						"module":  moduleName,
						"name":    viewDef.ID,
						"model":   "sys.view",
						"core_id": id,
					}, "name")
				}
			}
			continue
		}

		menus, err := parser.ParseMenus(xmlPath)
		if err == nil && len(menus) > 0 {
			for _, menu := range menus {
				menuValues := map[string]interface{}{
					"name":          menu.Name,
					"sequence":      menu.Sequence,
					"module":        moduleName,
					"access_groups": strings.TrimSpace(menu.AccessGroups),
				}

				if menu.Action != "" {
					actionID, err := resolveXMLIDInModule(context, moduleName, menu.Action)
					if err == nil && actionID != 0 {
						menuValues["action_id"] = actionID
					}
				}

				if sanitizedIcon := sanitizeWebIcon(menu.WebIcon); sanitizedIcon != "" {
					menuValues["web_icon"] = sanitizedIcon
				}

				if menu.ParentID != "" {
					parentID, err := resolveXMLIDInModule(context, moduleName, menu.ParentID)
					if err == nil && parentID != 0 {
						menuValues["parent_id"] = parentID
					}
				}

				id, err := orm.Upsert(context, orm.SysMenu{}, menuValues, "name")
				if err == nil {
					orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
						"module":  moduleName,
						"name":    menu.ID,
						"model":   "sys.menu",
						"core_id": id,
					}, "name")
				}
			}
		}
	}

	for _, xmlRecord := range inheritQueue {
		if err := applySysUIViewInherit(context, moduleName, xmlRecord); err != nil {
			fmt.Printf(platformmsg.FmtViewInheritWarning, moduleName, xmlRecord.ID, err)
		}
	}

	return nil
}

func syncGenericRegistryRecord(context context.Context, moduleName string, xmlRecord parser.Record) {
	if xmlRecord.Model == "sys.action.window" || xmlRecord.Model == "sys.view" {
		return
	}
	// Other sys.* models may be registered (sys.access, sys.rule, …).
	if strings.HasPrefix(xmlRecord.Model, "sys.") {
		modelInstance, ok := orm.Registry[xmlRecord.Model]
		if !ok || modelInstance == nil {
			return
		}
		syncRegistryRecordByModel(context, moduleName, xmlRecord, modelInstance)
		return
	}
	modelInstance, ok := orm.Registry[xmlRecord.Model]
	if !ok || modelInstance == nil {
		return
	}
	syncRegistryRecordByModel(context, moduleName, xmlRecord, modelInstance)
}

func syncRegistryRecordByModel(context context.Context, moduleName string, xmlRecord parser.Record, modelInstance orm.Model) {
	fieldMapStrings := parser.RecordFieldMap(xmlRecord)
	if len(fieldMapStrings) == 0 {
		return
	}
	fieldValues := map[string]interface{}{}
	for key, val := range fieldMapStrings {
		fieldValues[key] = convertRecordScalar(context, moduleName, xmlRecord.Model, key, val)
	}
	conflictColumn := "name"
	if xmlRecord.Model == "core.user" {
		conflictColumn = "login"
	}
	if _, ok := fieldValues[conflictColumn]; !ok {
		return
	}
	id, err := orm.Upsert(context, modelInstance, fieldValues, conflictColumn)
	if err != nil {
		fmt.Printf(platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
		return
	}
	_, _ = orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
		"module":  moduleName,
		"name":    xmlRecord.ID,
		"model":   xmlRecord.Model,
		"core_id": id,
	}, "name")
}

func (addon *Addon) syncCSVModelAccess(context context.Context) error {
	csvPath := filepath.Join(addon.Path, "sys.access.csv")
	if _, err := os.Stat(csvPath); err != nil {
		csvPath = filepath.Join(addon.Path, "security", "sys.access.csv")
		if _, err := os.Stat(csvPath); err != nil {
			return nil // No CSV ACL file found
		}
	}

	csvFile, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	// Skip header: id,name,model_id:id,group_id:id,perm_read,perm_write,perm_create,perm_unlink
	if _, err := csvReader.Read(); err != nil {
		return err
	}

	for {
		csvRecord, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(csvRecord) < 8 {
			continue
		}

		recordXmlId := csvRecord[0]
		accessName := csvRecord[1]
		modelName := csvRecord[2]
		groupXmlId := csvRecord[3]
		permRead := csvRecord[4] == "1"
		permWrite := csvRecord[5] == "1"
		permCreate := csvRecord[6] == "1"
		permUnlink := csvRecord[7] == "1"

		var groupId int
		if groupXmlId != "" {
			gid, _, err := orm.ResolveXmlId(context, groupXmlId)
			if err != nil {
				// Try with module prefix if not absolute
				if !strings.Contains(groupXmlId, ".") {
					gid, _, _ = orm.ResolveXmlId(context, addon.Manifest.Name+"."+groupXmlId)
				}
			}
			groupId = gid
		}

		accessValues := map[string]interface{}{
			"name":        accessName,
			"model":       modelName,
			"perm_read":   permRead,
			"perm_write":  permWrite,
			"perm_create": permCreate,
			"perm_unlink": permUnlink,
		}
		if groupId > 0 {
			accessValues["group_id"] = groupId
		}

		id, err := orm.Upsert(context, orm.SysAccess{}, accessValues, "name")
		if err == nil {
			orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
				"module": addon.Manifest.Name,
				"name":   recordXmlId,
				"model":  "sys.access",
				"res_id": id,
			}, "name")
		}
	}
	return nil
}

// viewArchXML persists the full parsed view (header, sheet, notebook, etc.) for sys.view.arch.
func viewArchXML(viewDef *parser.View) string {
	if viewDef == nil {
		return "<view/>"
	}
	marshaledXml, err := xml.Marshal(viewDef)
	if err != nil {
		// Minimal fallback so DB still gets a row
		return fmt.Sprintf("<view model=\"%s\" type=\"%s\"></view>", viewDef.Model, viewDef.Type)
	}
	return string(marshaledXml)
}

func sanitizeWebIcon(iconString string) string {
	iconString = strings.TrimSpace(iconString)
	if iconString == "" {
		return ""
	}
	for _, char := range iconString {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return ""
		}
	}
	return iconString
}

func convertRecordScalar(context context.Context, moduleName, model, column, rawValue string) interface{} {
	trimmedValue := strings.TrimSpace(rawValue)
	if strings.HasPrefix(column, "perm_") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	if column == "group_id" || column == "user_id" || column == "rule_id" || column == "implied_group_id" || column == "parent_id" {
		if trimmedValue == "" || strings.EqualFold(trimmedValue, "false") || trimmedValue == "0" {
			return nil
		}
		if strings.Contains(trimmedValue, ".") {
			if id, _, err := orm.ResolveXmlId(context, trimmedValue); err == nil && id > 0 {
				return id
			}
		}
		if moduleName != "" {
			if id, _, err := orm.ResolveXmlId(context, moduleName+"."+trimmedValue); err == nil && id > 0 {
				return id
			}
		}
	}
	if column == "active" || strings.HasSuffix(column, "_active") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	return trimmedValue
}
