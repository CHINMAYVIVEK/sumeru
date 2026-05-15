package module

import (
	"context"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

func syncMenusFromItems(ctx context.Context, moduleName string, menus []parser.MenuItem) {
	for _, menu := range menus {
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

		if menu.ParentID != "" {
			parentID, err := resolveXMLIDInModule(ctx, moduleName, menu.ParentID)
			if err == nil && parentID != 0 {
				menuValues["parent_id"] = parentID
			}
		}

		id, err := orm.Upsert(ctx, orm.SysMenu{}, menuValues, "name")
		if err == nil {
			_, _ = orm.Upsert(ctx, orm.SysModelData{}, map[string]interface{}{
				"module":  moduleName,
				"name":    menu.ID,
				"model":   "sys.menu",
				"core_id": id,
			}, "name")
		}
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
