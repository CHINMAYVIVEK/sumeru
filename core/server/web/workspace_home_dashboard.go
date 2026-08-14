package web

import (
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

type homeDashData struct {
	MenuID   string
	Search   string
	Tiles    []render.AppTile
	EmptyMsg string
}

// HomeDashboardHandler shows installed application modules as the signed-in user’s app hub.
func HomeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	searchQ := strings.TrimSpace(r.URL.Query().Get("q"))
	menuIDStr := strings.TrimSpace(r.URL.Query().Get("menu_id"))
	if menuIDStr == "" {
		if hid, _, err := orm.ResolveXmlId(ctx, "base.menu_home_root"); err == nil && hid > 0 {
			http.Redirect(w, r, fmt.Sprintf("/web/home?menu_id=%d", hid), http.StatusFound)
			return
		}
	}

	raw, err := loadInstalledAppTiles(ctx, searchQ)
	if err != nil {
		WebLogEvent(ctx, r.URL.Path, "Failed to list modules for home dashboard", "load", "failure", err, nil)
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return
	}

	tiles := raw

	emptyMsg := ""
	if len(tiles) == 0 {
		emptyMsg = "No installed applications match your search. Install apps from Apps or clear the search."
	}

	activeMenu := menuIDStr
	if activeMenu == "" {
		if hid, _, err := orm.ResolveXmlId(ctx, "base.menu_home_root"); err == nil {
			activeMenu = fmt.Sprintf("%d", hid)
		}
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         r.URL.Path,
		InnerTemplate: "home_dashboard_inner.html",
		InnerData: homeDashData{
			MenuID: menuIDStr, Search: searchQ, Tiles: tiles, EmptyMsg: emptyMsg,
		},
		MenuIDStr: activeMenu,
		Page: render.PageData{
			Title:                "Home",
			ViewBreadcrumb:       "Dashboard",
			ActiveMenuID:         activeMenu,
			ViewStylesheetURLs:   []string{"/static/css/sumeru-home.css"},
			SuppressActivityDock: true,
			BreadcrumbItems:      render.BuildHomeDashboardBreadcrumbs(ctx),
		},
	})
	WebLogNavigation(ctx, r.URL.Path, "module_hub", "Home dashboard opened", map[string]interface{}{
		"menu_id":    menuIDStr,
		"tile_count": len(tiles),
		"search":     searchQ,
	})
}

// homeSearchMatches returns true if every whitespace-separated token in q appears
// in name, display name, or description (case-insensitive).
func homeSearchMatches(q, technicalName, displayName, description string) bool {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return true
	}
	hay := strings.ToLower(strings.TrimSpace(technicalName) + " " + strings.TrimSpace(displayName) + " " + strings.TrimSpace(description))
	for _, tok := range strings.Fields(q) {
		t := strings.TrimSpace(strings.ToLower(tok))
		if t == "" {
			continue
		}
		if !strings.Contains(hay, t) {
			return false
		}
	}
	return true
}
