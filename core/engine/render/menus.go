package render

import (
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

var menuIconKey = regexp.MustCompile(`^[a-z0-9-]+$`)

func sanitizeMenuIcon(s string) string {
	s = strings.TrimSpace(s)
	if menuIconKey.MatchString(s) {
		return s
	}
	return ""
}

// LoadShellMenus loads ir.ui.menu rows for installed modules and derives top bar + sidebar structure.
// Top bar: one entry per installed application module (ir.module.application = true), ordered by each
// root menuitem's sequence (then name), like Odoo XML. The Settings app root (company module) is
// always placed second-to-last; the Apps link is rendered after TopMenus in base.html (always last).
// Sidebar: menus under the active app root (may span modules, e.g. Settings).
func LoadShellMenus(activeMenuID string) (topMenus []parser.MenuItem, sidebarMenus []SidebarMenu, activeModuleID string) {
	modTbl := orm.GetTableName("ir.module")
	menuTbl := orm.GetTableName("ir.ui.menu")
	query := fmt.Sprintf(
		`SELECT m.id, m.name, m.parent_id, m.action_id, m.sequence,
		        COALESCE(NULLIF(TRIM(m.module), ''), '') AS module,
		        COALESCE(NULLIF(TRIM(m.web_icon), ''), '') AS web_icon
		   FROM %s m
		  WHERE COALESCE(NULLIF(TRIM(m.module), ''), '') = ''
		     OR EXISTS (
		      SELECT 1 FROM %s im WHERE im.name = m.module AND im.state = 'installed' AND im.active = true
		    )
		  ORDER BY m.sequence`,
		menuTbl, modTbl,
	)
	rows, err := orm.DB.Query(query)
	if err != nil {
		log.Printf("Error fetching menus: %v", err)
		return nil, nil, ""
	}
	defer rows.Close()

	var allMenus []parser.MenuItem
	for rows.Next() {
		var id int
		var name, mod, webIcon string
		var parentID, actionID, seq sql.NullInt64
		if err := rows.Scan(&id, &name, &parentID, &actionID, &seq, &mod, &webIcon); err != nil {
			log.Printf("Error scanning menu row: %v", err)
			continue
		}

		m := parser.MenuItem{
			ID:       fmt.Sprintf("%d", id),
			Name:     name,
			Sequence: int(seq.Int64),
			Module:   mod,
			WebIcon:  sanitizeMenuIcon(webIcon),
			Action:   fmt.Sprintf("/web?menu_id=%d", id),
		}
		if parentID.Valid {
			m.ParentID = fmt.Sprintf("%d", parentID.Int64)
		}
		if actionID.Valid && actionID.Int64 != 0 {
			m.Action = fmt.Sprintf("/web?action=%d&menu_id=%d", actionID.Int64, id)
		}
		allMenus = append(allMenus, m)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Menu rows error: %v", err)
	}

	appMods, err := queryInstalledApplicationNames()
	if err != nil {
		log.Printf("installed application modules: %v", err)
	}

	for _, m := range allMenus {
		if m.ParentID != "" {
			continue
		}
		mod := strings.TrimSpace(m.Module)
		if len(appMods) > 0 {
			_, ok := appMods[mod]
			if mod == "" || !ok {
				continue
			}
		}
		topMenus = append(topMenus, m)
	}
	if len(topMenus) == 0 {
		for _, m := range allMenus {
			if m.ParentID == "" {
				topMenus = append(topMenus, m)
			}
		}
	}
	topMenus = sortTopBarRootMenus(topMenus)

	if activeMenuID == "" && len(topMenus) > 0 {
		activeModuleID = topMenus[0].ID
	} else {
		for _, m := range allMenus {
			if m.ID == activeMenuID {
				curr := m
				for curr.ParentID != "" {
					found := false
					for _, parent := range allMenus {
						if parent.ID == curr.ParentID {
							curr = parent
							found = true
							break
						}
					}
					if !found || curr.ParentID == "" {
						break
					}
				}
				activeModuleID = curr.ID
				break
			}
		}
	}

	if activeModuleID != "" {
		// Include any menu row under the active app root (may mix modules, e.g. Settings: company + user).
		menuUnderActiveRoot := func(mi parser.MenuItem) bool {
			curr := mi
			for i := 0; i <= len(allMenus); i++ {
				if curr.ParentID == "" {
					return curr.ID == activeModuleID
				}
				found := false
				for _, p := range allMenus {
					if p.ID == curr.ParentID {
						curr = p
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
			return false
		}
		for _, m := range allMenus {
			if m.ParentID != activeModuleID || !menuUnderActiveRoot(m) {
				continue
			}
			sm := SidebarMenu{ID: m.ID, Name: m.Name, Sequence: m.Sequence}
			for _, sub := range allMenus {
				if sub.ParentID != m.ID || !menuUnderActiveRoot(sub) {
					continue
				}
				sm.SubMenus = append(sm.SubMenus, sub)
			}
			sort.Slice(sm.SubMenus, func(i, j int) bool {
				if sm.SubMenus[i].Sequence != sm.SubMenus[j].Sequence {
					return sm.SubMenus[i].Sequence < sm.SubMenus[j].Sequence
				}
				return sm.SubMenus[i].Name < sm.SubMenus[j].Name
			})
			sidebarMenus = append(sidebarMenus, sm)
		}
		sort.Slice(sidebarMenus, func(i, j int) bool {
			if sidebarMenus[i].Sequence != sidebarMenus[j].Sequence {
				return sidebarMenus[i].Sequence < sidebarMenus[j].Sequence
			}
			return sidebarMenus[i].Name < sidebarMenus[j].Name
		})
	}

	return topMenus, sidebarMenus, activeModuleID
}

// sortTopBarRootMenus orders application root menus by each menuitem's `sequence` (then `name`),
// matching Odoo XML. The Settings root (company module, name "Settings") is pinned second-to-last;
// the Apps link is rendered after TopMenus in base.html and is always last.
func sortTopBarRootMenus(in []parser.MenuItem) []parser.MenuItem {
	if len(in) == 0 {
		return in
	}
	var settings *parser.MenuItem
	rest := make([]parser.MenuItem, 0, len(in))
	for i := range in {
		m := in[i]
		if isPinnedSettingsRootMenu(m) {
			mm := m
			settings = &mm
			continue
		}
		rest = append(rest, m)
	}
	sort.Slice(rest, func(i, j int) bool {
		if rest[i].Sequence != rest[j].Sequence {
			return rest[i].Sequence < rest[j].Sequence
		}
		return rest[i].Name < rest[j].Name
	})
	if settings == nil {
		return rest
	}
	out := append([]parser.MenuItem{}, rest...)
	out = append(out, *settings)
	return out
}

func isPinnedSettingsRootMenu(m parser.MenuItem) bool {
	if m.ParentID != "" {
		return false
	}
	return strings.TrimSpace(m.Module) == "company" && strings.TrimSpace(m.Name) == "Settings"
}

// ModuleNameForTopMenu returns the display name of the active root menu.
func ModuleNameForTopMenu(topMenus []parser.MenuItem, activeModuleID string) string {
	moduleName := AppDisplayName
	for _, m := range topMenus {
		if m.ID == activeModuleID {
			moduleName = m.Name
			break
		}
	}
	return moduleName
}

func queryInstalledApplicationNames() (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if orm.DB == nil {
		return out, fmt.Errorf("database not initialized")
	}
	q := `SELECT name FROM ` + orm.GetTableName("ir.module") + ` WHERE state = 'installed' AND active = true AND application = true`
	rows, err := orm.DB.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		out[strings.TrimSpace(n)] = struct{}{}
	}
	return out, rows.Err()
}
