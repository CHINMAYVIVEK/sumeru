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
// Top bar: one entry per installed application module (ir.module.application = true).
// Sidebar: menus for the active app only (same ir.ui.menu.module as the active root).
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
	sort.Slice(topMenus, func(i, j int) bool {
		if topMenus[i].Sequence != topMenus[j].Sequence {
			return topMenus[i].Sequence < topMenus[j].Sequence
		}
		return topMenus[i].Name < topMenus[j].Name
	})

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
		var rootMod string
		for _, m := range allMenus {
			if m.ID == activeModuleID {
				rootMod = strings.TrimSpace(m.Module)
				break
			}
		}
		menuMatchesModule := func(mi parser.MenuItem) bool {
			mm := strings.TrimSpace(mi.Module)
			if rootMod == "" {
				return mm == ""
			}
			return mm == rootMod
		}
		for _, m := range allMenus {
			if m.ParentID != activeModuleID || !menuMatchesModule(m) {
				continue
			}
			sm := SidebarMenu{ID: m.ID, Name: m.Name}
			for _, sub := range allMenus {
				if sub.ParentID != m.ID || !menuMatchesModule(sub) {
					continue
				}
				sm.SubMenus = append(sm.SubMenus, sub)
			}
			sidebarMenus = append(sidebarMenus, sm)
		}
	}

	return topMenus, sidebarMenus, activeModuleID
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
