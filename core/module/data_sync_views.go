package module

import (
	"context"
	"encoding/xml"
	"fmt"

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

func upsertInlineViewDef(ctx context.Context, moduleName string, viewDef *parser.View) {
	viewArchitecture := viewArchXML(viewDef)
	id, err := orm.Upsert(ctx, orm.SysView{}, map[string]interface{}{
		"name":  viewDef.Model + "." + viewDef.Type,
		"model": viewDef.Model,
		"type":  viewDef.Type,
		"arch":  viewArchitecture,
	}, "name")
	if err == nil {
		_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
			"module":  moduleName,
			"name":    viewDef.ID,
			"model":   "sys.view",
			"core_id": id,
		}, "name")
	}
}
