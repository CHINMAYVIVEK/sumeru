package module

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sumeru/core/base/platformmsg"
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
		if key == "implied_ids" {
			continue
		}
		fieldValues[key] = ConvertRecordScalar(ctx, moduleName, xmlRecord.Model, key, val)
	}
	conflictColumn := "name"
	if xmlRecord.Model == "core.user" {
		conflictColumn = "login"
	}
	if _, ok := fieldValues[conflictColumn]; !ok {
		return
	}
	id, err := orm.Upsert(ctx, modelInstance, fieldValues, conflictColumn)
	if err != nil {
		fmt.Printf(platformmsg.FmtGenericUpsertWarn, xmlRecord.Model, xmlRecord.ID, err)
		return
	}
	if xmlRecord.Model == "core.group" && impliedEval != "" {
		if err := syncCoreGroupImpliedFromEval(ctx, moduleName, id, impliedEval); err != nil {
			fmt.Printf("Warning: core.group implied_ids %s (%s): %v\n", xmlRecord.ID, moduleName, err)
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
	if column == "group_id" || column == "user_id" || column == "rule_id" || column == "implied_group_id" || column == "parent_id" || column == "category_id" {
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
	}
	if column == "active" || strings.HasSuffix(column, "_active") {
		if boolValue, err := strconv.ParseBool(trimmedValue); err == nil {
			return boolValue
		}
		return strings.EqualFold(trimmedValue, "true") || trimmedValue == "1"
	}
	return trimmedValue
}
