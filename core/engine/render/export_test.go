package render

import (
	"sumeru/core/engine/parser"
)

func BuildSidebarMenus(allMenus []parser.MenuItem, rootMenuID string, menuAllowed func(parser.MenuItem) bool) []SidebarMenu {
	return buildSidebarMenus(allMenus, rootMenuID, menuAllowed)
}

func ResolveActiveModuleID(allMenus []parser.MenuItem, activeMenuID string) string {
	return resolveActiveModuleID(allMenus, activeMenuID)
}
