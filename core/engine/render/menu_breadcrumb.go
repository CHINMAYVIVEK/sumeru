package render

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

type BreadcrumbItem struct {
	Label string
	Href  string // empty = current page (no link)
}

// HomeWebURL returns the canonical Home dashboard URL including menu_id when resolvable.
func HomeWebURL(ctx context.Context) string {
	id, _, err := orm.ResolveXmlId(ctx, "base.menu_home_root")
	if err != nil || id == 0 {
		return "/web/home"
	}
	return fmt.Sprintf("/web/home?menu_id=%d", id)
}

// SettingsHomeURL is the canonical Settings hub URL.
func SettingsHomeURL() string {
	return "/web/settings"
}

// MenuWebURL builds a /web URL for a sys.menu row.
func MenuWebURL(menuID, actionID int) string {
	if actionID > 0 {
		return fmt.Sprintf("/web?action=%d&menu_id=%d", actionID, menuID)
	}
	return fmt.Sprintf("/web?menu_id=%d", menuID)
}

// listViewURL returns the list view URL for a form workspace query string.
func listViewURL(formBaseQuery string) string {
	formBaseQuery = strings.TrimSpace(formBaseQuery)
	if formBaseQuery == "" {
		return ""
	}
	u, err := url.ParseQuery(formBaseQuery)
	if err != nil {
		return ""
	}
	u.Set("view_type", "list")
	u.Del("id")
	return "/web?" + u.Encode()
}

type menuCrumb struct {
	ID       int
	Name     string
	ActionID int
}

func collectMenuAncestors(ctx context.Context, leafID int) []menuCrumb {
	var stack []menuCrumb
	id := leafID
	seen := make(map[int]bool)
	for id > 0 && !seen[id] {
		seen[id] = true
		row, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": id})
		if err != nil {
			break
		}
		name := strings.TrimSpace(orm.AsString(row["name"]))
		aid, _ := orm.CoerceInt64(row["action_id"])
		stack = append(stack, menuCrumb{ID: id, Name: name, ActionID: int(aid)})
		pid := 0
		if pv, ok := row["parent_id"]; ok {
			if v, ok2 := orm.CoerceInt64(pv); ok2 && v > 0 {
				pid = int(v)
			}
		}
		id = pid
	}
	// leaf was pushed first; reverse to root→leaf
	for i, j := 0, len(stack)-1; i < j; i, j = i+1, j-1 {
		stack[i], stack[j] = stack[j], stack[i]
	}
	return stack
}

// BuildWorkspaceBreadcrumbs builds app-root menu path + current view/record for /web workspace pages.
// The trail starts at the active application root (same label as shell ModuleName in normal cases), not Home.
func BuildWorkspaceBreadcrumbs(ctx context.Context, activeMenuID string, viewType, viewHumanTitle, formBaseQuery string, record map[string]interface{}, recordID int) []BreadcrumbItem {
	var items []BreadcrumbItem

	mid, err := strconv.Atoi(strings.TrimSpace(activeMenuID))
	if err != nil || mid <= 0 {
		if strings.TrimSpace(viewHumanTitle) != "" {
			return []BreadcrumbItem{{Label: strings.TrimSpace(viewHumanTitle), Href: ""}}
		}
		return items
	}

	chain := collectMenuAncestors(ctx, mid)
	vt := strings.ToLower(strings.TrimSpace(viewType))
	isMatrix := vt == "list" || vt == "kanban"
	settingsRootID := 0
	if IsMenuUnderSettingsRoot(ctx, activeMenuID) {
		if rid, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root"); err == nil && rid > 0 {
			settingsRootID = rid
		}
	}

	for i, m := range chain {
		isLast := i == len(chain)-1
		href := MenuWebURL(m.ID, m.ActionID)
		if settingsRootID > 0 && m.ID == settingsRootID {
			href = SettingsHomeURL()
		}
		if isLast && isMatrix {
			href = ""
		}
		if isLast && vt == "form" && recordID > 0 {
			if listHref := listViewURL(formBaseQuery); listHref != "" {
				href = listHref
			}
		}
		items = append(items, BreadcrumbItem{Label: m.Name, Href: href})
	}

	if vt == "form" && recordID > 0 {
		label := strings.TrimSpace(viewHumanTitle)
		if record != nil {
			if n := strings.TrimSpace(recStr(record, "name")); n != "" {
				label = n
			}
		}
		if label == "" {
			label = "Record"
		}
		items = append(items, BreadcrumbItem{Label: label, Href: ""})
		return items
	}

	// List/kanban: menu chain already names the screen (e.g. "All Companies"); do not append model list title ("Companies").
	if isMatrix && len(chain) > 0 {
		return items
	}

	if strings.TrimSpace(viewHumanTitle) != "" {
		if len(items) == 0 || items[len(items)-1].Label != viewHumanTitle {
			items = append(items, BreadcrumbItem{Label: viewHumanTitle, Href: ""})
		}
	}
	return items
}

// BuildAppsBreadcrumbs returns Home + Apps (+ optional module detail as current).
func BuildAppsBreadcrumbs(ctx context.Context, appsListHref string, detailTitle string) []BreadcrumbItem {
	out := []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
	}
	if strings.TrimSpace(detailTitle) != "" {
		listHref := strings.TrimSpace(appsListHref)
		if listHref == "" {
			listHref = "/web/apps"
		}
		out = append(out, BreadcrumbItem{Label: "Apps", Href: listHref})
		out = append(out, BreadcrumbItem{Label: strings.TrimSpace(detailTitle), Href: ""})
		return out
	}
	out = append(out, BreadcrumbItem{Label: "Apps", Href: ""})
	return out
}

// BuildHomeDashboardBreadcrumbs returns Home + Dashboard (current).
func BuildHomeDashboardBreadcrumbs(ctx context.Context) []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
		{Label: "Dashboard", Href: ""},
	}
}

// BuildSettingsHubBreadcrumbs returns a single Settings crumb for the hub page.
func BuildSettingsHubBreadcrumbs(ctx context.Context) []BreadcrumbItem {
	return []BreadcrumbItem{
		{Label: "Settings", Href: ""},
	}
}

// BuildAppLogsBreadcrumbs returns Home + menu path to the Event Log menu (current = leaf).
func BuildAppLogsBreadcrumbs(ctx context.Context, appLogsMenuID int) []BreadcrumbItem {
	items := []BreadcrumbItem{
		{Label: "Home", Href: HomeWebURL(ctx)},
	}
	if appLogsMenuID <= 0 {
		return items
	}
	chain := collectMenuAncestors(ctx, appLogsMenuID)
	for i, m := range chain {
		isLast := i == len(chain)-1
		href := MenuWebURL(m.ID, m.ActionID)
		if isLast {
			href = ""
		}
		items = append(items, BreadcrumbItem{Label: m.Name, Href: href})
	}
	return items
}
