package render

import (
	"testing"

	"sumeru/core/engine/parser"
)

func TestBuildSidebarMenus_localizationSection(t *testing.T) {
	allMenus := []parser.MenuItem{
		{ID: "100", Name: "Settings", Sequence: 5},
		{ID: "160", Name: "Localization", ParentID: "100", Sequence: 60, AccessGroups: "base.group_system"},
		{ID: "161", Name: "Countries", ParentID: "160", Sequence: 10, Action: "/web?action=10&menu_id=161", AccessGroups: "base.group_system"},
		{ID: "162", Name: "States", ParentID: "160", Sequence: 20, Action: "/web?action=11&menu_id=162", AccessGroups: "base.group_system"},
		{ID: "163", Name: "Cities", ParentID: "160", Sequence: 30, Action: "/web?action=12&menu_id=163", AccessGroups: "base.group_system"},
	}
	menuAllowed := func(mi parser.MenuItem) bool {
		return mi.AccessGroups == "base.group_system"
	}
	sections := buildSidebarMenus(allMenus, "100", menuAllowed)
	var localization *SidebarMenu
	for i := range sections {
		if sections[i].Name == "Localization" {
			localization = &sections[i]
			break
		}
	}
	if localization == nil {
		t.Fatal("Localization section not found in Settings sidebar")
	}
	if len(localization.SubMenus) != 3 {
		t.Fatalf("Localization submenus = %d; want 3", len(localization.SubMenus))
	}
	names := map[string]bool{}
	for _, sm := range localization.SubMenus {
		names[sm.Name] = true
		if sm.Action == "" {
			t.Fatalf("submenu %q has empty action URL", sm.Name)
		}
	}
	for _, want := range []string{"Countries", "States", "Cities"} {
		if !names[want] {
			t.Fatalf("missing submenu %q", want)
		}
	}
}
