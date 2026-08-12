package module

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func syncGenericRegistryRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	if xmlRecord.Model == "sys.action.window" || xmlRecord.Model == "sys.view" {
		return
	}
	if strings.HasPrefix(xmlRecord.Model, "sys.") {
		modelInstance, ok := orm.Registry[xmlRecord.Model]
		if !ok || modelInstance == nil {
			return
		}
		syncRegistryRecordByModel(ctx, moduleName, xmlRecord, modelInstance)
		return
	}
	modelInstance, ok := orm.Registry[xmlRecord.Model]
	if !ok || modelInstance == nil {
		return
	}
	syncRegistryRecordByModel(ctx, moduleName, xmlRecord, modelInstance)
}

func syncRegistryRecordByModel(ctx context.Context, moduleName string, xmlRecord parser.Record, modelInstance orm.Model) {
	fieldMapStrings := parser.RecordFieldMap(xmlRecord)
	if len(fieldMapStrings) == 0 {
		return
	}
	impliedEval := strings.TrimSpace(fieldMapStrings["implied_ids"])
	fieldValues := map[string]interface{}{}
	for key, val := range fieldMapStrings {
		if key == "implied_ids" || key == "groups" {
			continue
		}
		fieldValues[key] = ConvertRecordScalar(ctx, moduleName, xmlRecord.Model, key, val)
	}

	// Prefer external id when already synced (states/cities have no natural unique key).
	if existingID, _, err := orm.ResolveXmlId(ctx, moduleName+"."+xmlRecord.ID); err == nil && existingID > 0 {
		if err := orm.UpdateRecordByID(ctx, xmlRecord.Model, existingID, fieldValues); err == nil {
			return
		}
		// Stale xml id — fall through to create and refresh model_data.
	}

	conflictColumn := "name"
	switch xmlRecord.Model {
	case "core.user":
		conflictColumn = "login"
	case "core.country", "core.lang":
		conflictColumn = "code"
	case "core.country.state", "core.city":
		conflictColumn = ""
	}

	// States/cities: match existing row by name + country (and state for cities) so -u
	// after deleteModuleMetadata does not insert duplicates.
	if xmlRecord.Model == "core.country.state" || xmlRecord.Model == "core.city" {
		criteria := map[string]interface{}{"name": fieldValues["name"]}
		if cid, ok := fieldValues["country_id"]; ok && cid != nil {
			criteria["country_id"] = cid
		}
		if xmlRecord.Model == "core.city" {
			if sid, ok := fieldValues["state_id"]; ok && sid != nil {
				criteria["state_id"] = sid
			}
		}
		if existing, err := orm.SearchOne(ctx, xmlRecord.Model, criteria); err == nil {
			if eid, ok := orm.CoerceInt64(existing["id"]); ok && eid > 0 {
				if err := orm.UpdateRecordByID(ctx, xmlRecord.Model, int(eid), fieldValues); err != nil {
					fmt.Printf(platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
					return
				}
				_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
					"module":  moduleName,
					"name":    xmlRecord.ID,
					"model":   xmlRecord.Model,
					"core_id": int(eid),
				}, "name")
				return
			}
		}
	}

	var id int
	var err error
	if conflictColumn == "" {
		id, err = orm.Create(ctx, modelInstance, fieldValues)
	} else {
		if _, ok := fieldValues[conflictColumn]; !ok {
			return
		}
		id, err = orm.Upsert(ctx, modelInstance, fieldValues, conflictColumn)
	}
	if err != nil {
		fmt.Printf(platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
		return
	}
	if xmlRecord.Model == "core.group" {
		if impliedEval != "" {
			if err := syncCoreGroupImpliedFromEval(ctx, moduleName, id, impliedEval); err != nil {
				fmt.Printf("Warning: core.group implied_ids %s (%s): %v\n", xmlRecord.ID, moduleName, err)
			}
		}
		if err := EnsureSystemImpliesManagerGroup(ctx, moduleName, xmlRecord.ID, id); err != nil {
			fmt.Printf("Warning: system→manager imply %s (%s): %v\n", xmlRecord.ID, moduleName, err)
		}
	}
	if xmlRecord.Model == "sys.rule" {
		if groupsEval := strings.TrimSpace(fieldMapStrings["groups"]); groupsEval != "" {
			if err := syncSysRuleGroupsFromEval(ctx, moduleName, id, groupsEval); err != nil {
				fmt.Printf("Warning: sys.rule groups %s (%s): %v\n", xmlRecord.ID, moduleName, err)
			}
		}
	}
	_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
		"module":  moduleName,
		"name":    xmlRecord.ID,
		"model":   xmlRecord.Model,
		"core_id": id,
	}, "name")
}

// ConvertRecordScalar coerces XML/form string values into types used for registry upserts.
func ConvertRecordScalar(ctx context.Context, moduleName, model, column, rawValue string) interface{} {
	trimmedValue := strings.TrimSpace(rawValue)
	if strings.HasPrefix(column, "perm_") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	if column == "group_id" || column == "user_id" || column == "rule_id" || column == "implied_group_id" || column == "parent_id" || column == "category_id" || column == "country_id" || column == "state_id" || column == "city_id" || strings.HasSuffix(column, "_id") {
		if trimmedValue == "" || strings.EqualFold(trimmedValue, "false") || trimmedValue == "0" {
			return nil
		}
		if strings.Contains(trimmedValue, ".") {
			if id, _, err := orm.ResolveXmlId(ctx, trimmedValue); err == nil && id > 0 {
				return id
			}
		}
		if moduleName != "" {
			if id, _, err := orm.ResolveXmlId(ctx, moduleName+"."+trimmedValue); err == nil && id > 0 {
				return id
			}
		}
		if n, err := strconv.ParseInt(trimmedValue, 10, 64); err == nil {
			return n
		}
		return nil
	}
	if column == "active" || strings.HasSuffix(column, "_active") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	return trimmedValue
}
