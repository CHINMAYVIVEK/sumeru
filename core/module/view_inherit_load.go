package module

import (
	"sumeru/core/orm"
	"sumeru/core/engine/parser"
	"sumeru/core/engine/viewinherit"
	"fmt"
	"strings"
)

// applyIrUIViewInherit merges an ir.ui.view inherit <record> into the parent view row (same DB id).
func applyIrUIViewInherit(modName string, rec parser.Record) error {
	fm := parser.RecordFieldMap(rec)
	ref := strings.TrimSpace(fm["inherit_id"])
	frag := fm["arch"]
	if ref == "" {
		return fmt.Errorf("inherit_id missing on record %q", rec.ID)
	}
	if strings.TrimSpace(frag) == "" {
		return fmt.Errorf("arch missing on inherit record %q", rec.ID)
	}
	parentID, _, err := orm.ResolveXmlId(ref)
	if err != nil || parentID == 0 {
		return fmt.Errorf("resolve inherit_id %q: %w", ref, err)
	}
	parent, err := orm.SearchOne("ir.ui.view", map[string]interface{}{"id": parentID})
	if err != nil {
		return fmt.Errorf("load parent view id %d: %w", parentID, err)
	}
	parentArch := orm.AsString(parent["arch"])
	if strings.TrimSpace(parentArch) == "" {
		return fmt.Errorf("parent view %d has empty arch", parentID)
	}
	merged, err := viewinherit.ApplyInheritArch(parentArch, frag)
	if err != nil {
		return fmt.Errorf("merge inherit %q: %w", rec.ID, err)
	}
	tbl := orm.GetTableName("ir.ui.view")
	if _, err := orm.DB.Exec(`UPDATE `+tbl+` SET arch = $1 WHERE id = $2`, merged, parentID); err != nil {
		return err
	}
	// Optional: map extension xml id to parent row for external id lookups
	if rec.ID != "" {
		if _, err := orm.Upsert(orm.IrModelData{}, map[string]interface{}{
			"module": modName,
			"name":   rec.ID,
			"model":  "ir.ui.view",
			"res_id": parentID,
		}, "name"); err != nil {
			return err
		}
	}
	return nil
}
