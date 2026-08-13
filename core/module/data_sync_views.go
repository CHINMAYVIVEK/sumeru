package module

import (
	"context"
	"sumeru/core/applog"
	"encoding/xml"
	"fmt"
	"strings"

	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// viewArchXML persists the full parsed view (header, sheet, notebook, etc.) for sys.view.arch.
func viewArchXML(viewDef *parser.View) string {
	if viewDef == nil {
		return "<view/>"
	}
	marshaledXml, err := xml.Marshal(viewDef)
	if err != nil {
		return fmt.Sprintf("<view model=\"%s\" type=\"%s\"></view>", viewDef.Model, viewDef.Type)
	}
	return string(marshaledXml)
}

// upsertSysViewFromRecord persists <record model="sys.view">…</record> data (non-inherit rows).
// Inline <view> elements use upsertInlineViewDef instead; inherit-only records are handled elsewhere.
func upsertSysViewFromRecord(ctx context.Context, moduleName string, xmlRecord parser.Record) {
	fieldMap := parser.RecordFieldMap(xmlRecord)
	recordValues := map[string]interface{}{}
	for key, val := range fieldMap {
		if key == "inherit_id" {
			continue
		}
		recordValues[key] = val
	}
	if _, ok := recordValues["name"]; !ok || strings.TrimSpace(orm.AsString(recordValues["name"])) == "" {
		recordValues["name"] = xmlRecord.ID
	}
	modelName := strings.TrimSpace(orm.AsString(recordValues["model"]))
	if modelName == "" {
		applog.L(context.Background()).Warn("module_sync", "msg", fmt.Sprintf("Warning: sys.view record %s (module %s): model is required", xmlRecord.ID, moduleName))
		return
	}
	arch := strings.TrimSpace(orm.AsString(recordValues["arch"]))
	if arch == "" {
		applog.L(context.Background()).Warn("module_sync", "msg", fmt.Sprintf("Warning: sys.view record %s (module %s): arch is required", xmlRecord.ID, moduleName))
		return
	}
	recordValues["arch"] = arch

	vt := strings.TrimSpace(strings.ToLower(orm.AsString(recordValues["type"])))
	if vt == "" {
		vt = inferSysViewTypeFromArch(arch)
	}
	if vt == "list" {
		vt = "tree"
	}
	if vt == "" {
		applog.L(context.Background()).Warn("module_sync", "msg", fmt.Sprintf("Warning: sys.view record %s (module %s): could not infer type from arch", xmlRecord.ID, moduleName))
		return
	}
	recordValues["type"] = vt

	id, err := orm.Upsert(ctx, orm.SysView{}, recordValues, "name")
	if err != nil {
		syncWarn(ctx, platformmsg.FmtGenericUpsertWarn, "sys.view", xmlRecord.ID, err)
		return
	}
	_ = linkXMLRecord(ctx, moduleName, xmlRecord.ID, "sys.view", id)
}

func inferSysViewTypeFromArch(arch string) string {
	a := strings.TrimSpace(arch)
	if a == "" {
		return ""
	}
	la := strings.ToLower(a)
	switch {
	case strings.HasPrefix(la, "<tree"):
		return "tree"
	case strings.HasPrefix(la, "<form"):
		return "form"
	case strings.HasPrefix(la, "<kanban"):
		return "kanban"
	case strings.HasPrefix(la, "<view"):
		if v, err := parser.ParseViewFromArch(a); err == nil {
			return strings.ToLower(strings.TrimSpace(v.Type))
		}
	}
	return ""
}

func upsertInlineViewDef(ctx context.Context, moduleName string, viewDef *parser.View) {
	viewArchitecture := viewArchXML(viewDef)
	id, err := orm.Upsert(ctx, orm.SysView{}, map[string]interface{}{
		"name":  viewDef.Model + "." + viewDef.Type,
		"model": viewDef.Model,
		"type":  viewDef.Type,
		"arch":  viewArchitecture,
	}, "name")
	if err == nil {
		_ = linkXMLRecord(ctx, moduleName, viewDef.ID, "sys.view", id)
	}
}
