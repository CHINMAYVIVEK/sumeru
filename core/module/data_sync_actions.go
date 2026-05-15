package module

import (
	"context"
	"fmt"
	"strings"

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
