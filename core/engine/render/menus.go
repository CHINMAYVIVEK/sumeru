package render

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
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

// LoadShellMenus loads sys.menu rows for installed modules and derives top bar + sidebar structure.
// Top bar: one entry per installed application module (sys.module.application = true), ordered by each
// root menuitem's sequence (then name), like Sumeru XML. A root named "Settings" is omitted here because
// base.html always renders a pinned link to /web/settings (second-to-last); Apps is always last.
// Sidebar: menus under the active app root (may span modules, e.g. Settings).
// shellModuleTitle is the active root menu label for breadcrumbs (works even when that root is not in TopMenus).
func LoadShellMenus(ctx context.Context, activeMenuID string) (topMenus []parser.MenuItem, sidebarMenus []SidebarMenu, activeModuleID, shellModuleTitle string) {
	shellModuleTitle = AppDisplayName
	modTbl := orm.GetTableName("sys.module")
	menuTbl := orm.GetTableName("sys.menu")
	query := fmt.Sprintf(
		`SELECT m.id, m.name, m.parent_id, m.action_id, m.sequence,
		        COALESCE(NULLIF(TRIM(m.module), ''), '') AS module,
		        COALESCE(NULLIF(TRIM(m.web_icon), ''), '') AS web_icon,
		        COALESCE(NULLIF(TRIM(m.access_groups), ''), '') AS access_groups
		   FROM %s m
		  WHERE COALESCE(NULLIF(TRIM(m.module), ''), '') = ''
		     OR EXISTS (
		      SELECT 1 FROM %s im WHERE im.name = m.module AND im.state = 'installed' AND im.active = true
		    )
		  ORDER BY m.sequence`,
		menuTbl, modTbl,
	)
	rows, err := orm.DB.QueryContext(ctx, query)
	if err != nil {
		log.Printf("Error fetching menus: %v", err)
		return nil, nil, "", AppDisplayName
	}
	defer rows.Close()

	var allMenus []parser.MenuItem
	for rows.Next() {
		var id int
		var name, mod, webIcon, accessGroups string
		var parentID, actionID, seq sql.NullInt64
		if err := rows.Scan(&id, &name, &parentID, &actionID, &seq, &mod, &webIcon, &accessGroups); err != nil {
			log.Printf("Error scanning menu row: %v", err)
			continue
		}

		lang := "en_US"
		if uid := orm.UIDFromContext(ctx); uid > 0 {
			if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
				if l := strings.TrimSpace(orm.AsString(u["lang"])); l != "" {
					lang = l
				}
			}
		}
		m := parser.MenuItem{
			ID:           fmt.Sprintf("%d", id),
			Name:         orm.Translate(ctx, lang, name),
			Sequence:     int(seq.Int64),
			Module:       mod,
			WebIcon:      sanitizeMenuIcon(webIcon),
			AccessGroups: strings.TrimSpace(accessGroups),
			Action:       fmt.Sprintf("/web?menu_id=%d", id),
		}
		// Treat self-referential parent_id (id = parent) as no parent — bad rows otherwise never
		// appear as top-bar roots or sidebar roots.
		if parentID.Valid && parentID.Int64 != int64(id) {
			m.ParentID = fmt.Sprintf("%d", parentID.Int64)
		}
		if actionID.Valid && actionID.Int64 != 0 {
			m.Action = fmt.Sprintf("/web?action=%d&menu_id=%d", actionID.Int64, id)
		} else if !parentID.Valid && strings.EqualFold(strings.TrimSpace(name), "Home") {
			m.Action = fmt.Sprintf("/web/home?menu_id=%d", id)
		}
		allMenus = append(allMenus, m)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Menu rows error: %v", err)
	}

	appMods, err := queryInstalledApplicationNames(ctx)
	if err != nil {
		log.Printf("installed application modules: %v", err)
	}

	uid := orm.UIDFromContext(ctx)
	menuAllowed := func(mi parser.MenuItem) bool {
		return orm.UserHasAnyAccessGroup(ctx, uid, mi.AccessGroups)
	}

	for _, m := range allMenus {
		if m.ParentID != "" {
			continue
		}
		if !menuAllowed(m) {
			continue
		}
		mod := strings.TrimSpace(m.Module)
		if len(appMods) > 0 {
			_, ok := appMods[mod]
			if mod == "" || !ok {
				continue
			}
		}
		// Pinned /web/settings in base.html — do not duplicate as a dynamic top item.
		if strings.EqualFold(strings.TrimSpace(m.Name), "Settings") {
			continue
		}
		// Home hub opens from the brand lockup (see base.html); do not duplicate in the top bar.
		if strings.EqualFold(strings.TrimSpace(m.Name), "Home") {
			continue
		}
		topMenus = append(topMenus, m)
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
		activeRootName := ""
		for _, m := range allMenus {
			if m.ID == activeModuleID {
				activeRootName = m.Name
				break
			}
		}
		if strings.TrimSpace(activeRootName) != "" {
			shellModuleTitle = activeRootName
		}

		// Collect direct children of the active root as sidebar section entries.
		// Each section may hold leaf links as its SubMenus ([]parser.MenuItem).
		var sections []SidebarMenu
		for _, m := range allMenus {
			if m.ParentID != activeModuleID || !menuAllowed(m) {
				continue
			}
			section := SidebarMenu{ID: m.ID, Name: m.Name, Sequence: m.Sequence}
			for _, sub := range allMenus {
				if sub.ParentID != m.ID || !menuAllowed(sub) {
					continue
				}
				section.SubMenus = append(section.SubMenus, sub)
			}
			sort.Slice(section.SubMenus, func(i, j int) bool {
				if section.SubMenus[i].Sequence != section.SubMenus[j].Sequence {
					return section.SubMenus[i].Sequence < section.SubMenus[j].Sequence
				}
				return section.SubMenus[i].Name < section.SubMenus[j].Name
			})
			sections = append(sections, section)
		}
		sort.Slice(sections, func(i, j int) bool {
			if sections[i].Sequence != sections[j].Sequence {
				return sections[i].Sequence < sections[j].Sequence
			}
			return sections[i].Name < sections[j].Name
		})
		sidebarMenus = append(sidebarMenus, sections...)
	}

	return topMenus, sidebarMenus, activeModuleID, shellModuleTitle
}

// sortTopBarRootMenus orders application root menus by each menuitem's `sequence` (then `name`),
// matching Sumeru XML.
func sortTopBarRootMenus(in []parser.MenuItem) []parser.MenuItem {
	if len(in) == 0 {
		return in
	}
	out := append([]parser.MenuItem{}, in...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Sequence != out[j].Sequence {
			return out[i].Sequence < out[j].Sequence
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func queryInstalledApplicationNames(ctx context.Context) (map[string]struct{}, error) {
	out := make(map[string]struct{})
	if orm.DB == nil {
		return out, fmt.Errorf("database not initialized")
	}
	q := `SELECT name FROM ` + orm.GetTableName("sys.module") + ` WHERE state = 'installed' AND active = true AND application = true`
	rows, err := orm.DB.QueryContext(ctx, q)
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

// IsMenuUnderSettingsRoot reports whether activeMenuID is the Settings root or one of its descendants.
func IsMenuUnderSettingsRoot(ctx context.Context, activeMenuID string) bool {
	id64, err := strconv.ParseInt(strings.TrimSpace(activeMenuID), 10, 64)
	if err != nil || id64 <= 0 || orm.DB == nil {
		return false
	}
	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || rootID == 0 {
		return false
	}
	cur := int(id64)
	for i := 0; i < 64; i++ {
		if cur == rootID {
			return true
		}
		row, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": cur})
		if err != nil {
			return false
		}
		pid, ok := orm.CoerceInt64(row["parent_id"])
		if !ok || pid == 0 {
			return false
		}
		cur = int(pid)
	}
	return false
}
