package module

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func syncMenusFromItems(ctx context.Context, moduleName string, menus []parser.MenuItem) {
	for _, menu := range menus {
		if strings.TrimSpace(menu.ID) == "" {
			continue
		}
		menuValues := map[string]interface{}{
			"name":          menu.Name,
			"sequence":      menu.Sequence,
			"module":        moduleName,
			"access_groups": strings.TrimSpace(menu.AccessGroups),
		}

		if menu.Action != "" {
			actionID, err := resolveXMLIDInModule(ctx, moduleName, menu.Action)
			if err == nil && actionID != 0 {
				menuValues["action_id"] = actionID
			}
		}

		if sanitizedIcon := sanitizeWebIcon(menu.WebIcon); sanitizedIcon != "" {
			menuValues["web_icon"] = sanitizedIcon
		}

		// Root menuitems have no parent in XML. Always persist NULL for parent_id so a later
		// update clears any stale value (e.g. self-referential parent_id == row id), which would
		// otherwise hide the menu from the top bar (roots are parent_id IS NULL / empty ParentID).
		if pid := strings.TrimSpace(menu.ParentID); pid != "" {
			parentID, err := resolveXMLIDInModule(ctx, moduleName, pid)
			if err == nil && parentID != 0 {
				menuValues["parent_id"] = parentID
			}
		} else {
			menuValues["parent_id"] = nil
		}

		// Prefer stable identity via sys.model_data (XML id), not Upsert on display name:
		// sys.menu.name is globally unique in the schema; name-based upsert corrupts sibling items.
		var rowID int
		if md, err := orm.SearchOne(ctx, "sys.model_data", map[string]interface{}{
			"module": moduleName,
			"model":  "sys.menu",
			"name":   menu.ID,
		}); err == nil {
			if cid, ok := orm.CoerceInt64(md["core_id"]); ok && cid > 0 {
				rowID = int(cid)
			}
		}
		if rowID > 0 {
			if err := orm.UpdateRecordByID(ctx, "sys.menu", rowID, menuValues); err != nil {
				fmt.Printf("Warning: sys.menu update %s.%s id=%d: %v\n", moduleName, menu.ID, rowID, err)
				continue
			}
		} else {
			id, err := orm.Create(ctx, orm.SysMenu{}, menuValues)
			if err != nil {
				fmt.Printf("Warning: sys.menu create %s.%s: %v\n", moduleName, menu.ID, err)
				continue
			}
			rowID = id
		}

		_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
			"module":  moduleName,
			"name":    menu.ID,
			"model":   "sys.menu",
			"core_id": rowID,
		}, "name")
	}
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
