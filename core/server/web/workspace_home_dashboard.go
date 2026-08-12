package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
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
		http.Error(w, fmt.Sprintf("Failed to list modules: %v", err), http.StatusInternalServerError)
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

	innerPath := filepath.Join(config.AppConfig.TemplatesPath, "home_dashboard_inner.html")
	tmpl, err := template.ParseFiles(innerPath)
	if err != nil {
		log.Printf("home dashboard: parse %s: %v", innerPath, err)
		http.Error(w, platformmsg.MsgHTTPTemplateError, http.StatusInternalServerError)
		return
	}
	var innerBuf bytes.Buffer
	if err := tmpl.Execute(&innerBuf, homeDashData{MenuID: menuIDStr, Search: searchQ, Tiles: tiles, EmptyMsg: emptyMsg}); err != nil {
		log.Printf("home dashboard: execute: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	activeMenu := menuIDStr
	if activeMenu == "" {
		if hid, _, err := orm.ResolveXmlId(ctx, "base.menu_home_root"); err == nil {
			activeMenu = fmt.Sprintf("%d", hid)
		}
	}
	topMenus, sidebarMenus, activeModuleID, moduleName := render.LoadShellMenus(ctx, activeMenu)
	page := render.PageData{
		Title:                "Home",
		ViewBreadcrumb:       "Dashboard",
		ModuleName:           moduleName,
		Content:              template.HTML(innerBuf.String()),
		TopMenus:             topMenus,
		SidebarMenus:         sidebarMenus,
		ActiveModuleID:       activeModuleID,
		ActiveMenuID:         activeMenu,
		ViewStylesheetURLs:   []string{"/static/css/sumeru-home.css"},
		ExtraStylesheetURLs:  render.ExtraStylesheetURLs,
		SuppressActivityDock: true,
		BreadcrumbItems:      render.BuildHomeDashboardBreadcrumbs(ctx),
	}

	html, err := render.RenderPage(ctx, config.AppConfig.TemplatesPath, page)
	if err != nil {
		http.Error(w, fmt.Sprintf("Layout: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
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
