package web

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"sumeru/core/sdk/platformmsg"
	"sumeru/core/engine/render"
	"sumeru/core/module"
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

	raw, err := module.ListModules(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list modules: %v", err), http.StatusInternalServerError)
		return
	}

	var tiles []dashAppTile
	for _, row := range raw {
		name := stringField(row["name"])
		if name == "" {
			continue
		}
		if !boolField(row["application"]) {
			continue
		}
		if stringField(row["state"]) != "installed" {
			continue
		}
		if !boolField(row["active"]) {
			continue
		}
		dn := stringField(row["display_name"])
		if dn == "" {
			dn = name
		}
		desc := stringField(row["description"])
		if searchQ != "" && !homeSearchMatches(searchQ, name, dn, desc) {
			continue
		}
		letter := "?"
		if r := []rune(strings.TrimSpace(dn)); len(r) > 0 {
			letter = strings.ToUpper(string(r[0]))
		}
		openHref := "/web/apps"
		if mid := rootMenuIDForModule(ctx, name); mid > 0 {
			openHref = fmt.Sprintf("/web?menu_id=%d", mid)
		}
		tiles = append(tiles, dashAppTile{
			Name:         name,
			DisplayName:  dn,
			Version:      stringField(row["version"]),
			Description:  desc,
			Author:       stringField(row["author"]),
			IconLetter:   letter,
			IconHue:      iconHueFromString(name),
			OpenMenuHref: openHref,
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

func rootMenuIDForModule(ctx context.Context, moduleName string) int {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return 0
	}
	tbl := orm.GetTableName("sys.menu")
	q := `SELECT id FROM ` + tbl + ` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`
	var id int
	if err := orm.DB.QueryRowContext(ctx, q, strings.TrimSpace(moduleName)).Scan(&id); err != nil {
		return 0
	}
	return id
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

func iconHueFromString(s string) int {
	h := 265
	for _, c := range strings.TrimSpace(s) {
		h = (h*31 + int(c)) % 360
		if h < 0 {
			h += 360
		}
	}
	return h
}
