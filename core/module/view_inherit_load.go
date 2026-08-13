package module

import (
	"context"
	"fmt"
	"strings"
	"sumeru/core/engine/parser"
	"sumeru/core/engine/viewinherit"
	"sumeru/core/orm"
)

// applySysUIViewInherit merges an sys.view inherit <record> into the parent view row (same DB id).
func applySysUIViewInherit(context context.Context, moduleName string, xmlRecord parser.Record) error {
	fieldMap := parser.RecordFieldMap(xmlRecord)
	inheritReference := strings.TrimSpace(fieldMap["inherit_id"])
	architectureFragment := fieldMap["arch"]
	if inheritReference == "" {
		return fmt.Errorf("inherit_id missing on record %q", xmlRecord.ID)
	}
	if strings.TrimSpace(architectureFragment) == "" {
		return fmt.Errorf("arch missing on inherit record %q", xmlRecord.ID)
	}
	parentID, err := resolveXMLIDInModule(context, moduleName, inheritReference)
	if err != nil || parentID == 0 {
		return fmt.Errorf("resolve inherit_id %q: %w", inheritReference, err)
	}
	parentView, err := orm.SearchOne(context, "sys.view", map[string]interface{}{"id": parentID})
	if err != nil {
		return fmt.Errorf("load parent view id %d: %w", parentID, err)
	}
	parentArchitecture := orm.AsString(parentView["arch"])
	if strings.TrimSpace(parentArchitecture) == "" {
		return fmt.Errorf("parent view %d has empty arch", parentID)
	}
	mergedArchitecture, err := viewinherit.ApplyInheritArch(parentArchitecture, architectureFragment)
	if err != nil {
		return fmt.Errorf("merge inherit %q: %w", xmlRecord.ID, err)
	}
	viewTableName := orm.MustQuotedTableName("sys.view")
	if _, err := orm.DB.ExecContext(context, `UPDATE `+viewTableName+` SET arch = $1 WHERE id = $2`, mergedArchitecture, parentID); err != nil {
		return err
	}
	// Optional: map extension xml id to parent row for external id lookups
	if xmlRecord.ID != "" {
		if _, err := orm.Upsert(context, orm.SysModelData{}, map[string]interface{}{
			"module":  moduleName,
			"name":    xmlRecord.ID,
			"model":   "sys.view",
			"core_id": parentID,
		}, "name"); err != nil {
			return err
		}
	}
	return nil
}
