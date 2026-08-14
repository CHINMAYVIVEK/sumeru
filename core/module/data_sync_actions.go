package module

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func upsertSysActionWindowFromRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	recordValues := recordValuesFromXML(xmlRecord)
	if cm := strings.TrimSpace(orm.AsString(recordValues["core_model"])); cm == "" {
		syncWarn(ctx, "Warning: sys.action.window record %s (module %s): core_model is required", xmlRecord.ID, moduleName)
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
