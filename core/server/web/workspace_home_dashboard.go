package web

import (
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

type dashAppTile struct {
	Name         string
	DisplayName  string
	Version      string
	Description  string
	Author       string
	IconLetter   string
	IconHue      int    // 0–359 for per-app icon tint (HSL hue)
	OpenMenuHref string // /web?menu_id=… or /web/apps fallback
}

type homeDashData struct {
	MenuID   string
	Search   string
	Tiles    []dashAppTile
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

	var tiles []dashAppTile
	for _, t := range raw {
		tiles = append(tiles, dashAppTile{
			Name:         t.Name,
			DisplayName:  t.DisplayName,
			Version:      t.Version,
			Description:  t.Description,
			Author:       t.Author,
			IconLetter:   t.IconLetter,
			IconHue:      t.IconHue,
			OpenMenuHref: t.OpenMenuHref,
		})
	}

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
