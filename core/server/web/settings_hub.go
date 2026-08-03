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
	"sumeru/core/module"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// settingsHubLink is one entry under a settings section.
type settingsHubLink struct {
	Name string
	Href string
}

// settingsHubSection groups links (e.g. Companies, Users) from the Settings menu tree.
type settingsHubSection struct {
	Title      string
	FilterText string
	Links      []settingsHubLink
}

// settingsHubAppTile is an installed application module with a quick open link.
type settingsHubAppTile struct {
	Name         string
	DisplayName  string
	IconLetter   string
	IconHue      int
	OpenMenuHref string
}

type settingsHubData struct {
	Sections          []settingsHubSection
	AppTiles          []settingsHubAppTile
	CompaniesMenuHref string
}

// SettingsHubHandler renders the default /web/settings overview: sections from the Settings
// menu tree plus installed application modules for quick access.
func SettingsHubHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()

	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || rootID == 0 {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
		return
	}
	menuIDStr := fmt.Sprintf("%d", rootID)

	topMenus, sidebarMenus, activeModuleID, moduleName := render.LoadShellMenus(ctx, menuIDStr)

	var sections []settingsHubSection
	for _, sec := range sidebarMenus {
		var links []settingsHubLink
		var ftParts []string
		ftParts = append(ftParts, strings.ToLower(strings.TrimSpace(sec.Name)))
		for _, sm := range sec.SubMenus {
			name := strings.TrimSpace(sm.Name)
			href := strings.TrimSpace(sm.Action)
			if name == "" || href == "" {
				continue
			}
			links = append(links, settingsHubLink{Name: name, Href: href})
			ftParts = append(ftParts, strings.ToLower(name))
		}
		if len(links) == 0 {
			continue
		}
		sections = append(sections, settingsHubSection{
			Title:      strings.TrimSpace(sec.Name),
			FilterText: strings.Join(ftParts, " "),
			Links:      links,
		})
	}

	raw, err := module.ListModules(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list modules: %v", err), http.StatusInternalServerError)
		return
	}
	var appTiles []settingsHubAppTile
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
		letter := "?"
		if ru := []rune(strings.TrimSpace(dn)); len(ru) > 0 {
			letter = strings.ToUpper(string(ru[0]))
		}
		openHref := "/web/apps"
		if mid := rootMenuIDForModule(ctx, name); mid > 0 {
			openHref = fmt.Sprintf("/web?menu_id=%d", mid)
		}
		appTiles = append(appTiles, settingsHubAppTile{
			Name:         name,
			DisplayName:  dn,
			IconLetter:   letter,
			IconHue:      iconHueFromString(name),
			OpenMenuHref: openHref,
		})
	}

	companiesHref := ""
	if mid, _, err := orm.ResolveXmlId(ctx, "base.menu_company_companies"); err == nil && mid > 0 {
		companiesHref = fmt.Sprintf("/web?menu_id=%d", mid)
	}

	innerPath := filepath.Join(config.AppConfig.TemplatesPath, "settings_hub_inner.html")
	tmpl, err := template.ParseFiles(innerPath)
	if err != nil {
		log.Printf("settings hub: parse %s: %v", innerPath, err)
		http.Error(w, platformmsg.MsgHTTPTemplateError, http.StatusInternalServerError)
		return
	}
	var innerBuf bytes.Buffer
	data := settingsHubData{
		Sections:          sections,
		AppTiles:          appTiles,
		CompaniesMenuHref: companiesHref,
	}
	if err := tmpl.Execute(&innerBuf, data); err != nil {
		log.Printf("settings hub: execute: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	page := render.PageData{
		Title:                "Settings",
		ViewBreadcrumb:       "Overview",
		ModuleName:           moduleName,
		Content:              template.HTML(innerBuf.String()),
		TopMenus:             topMenus,
		SidebarMenus:         sidebarMenus,
		ActiveModuleID:       activeModuleID,
		ActiveMenuID:         menuIDStr,
		SettingsNavActive:    true,
		SuppressActivityDock: true,
		BreadcrumbItems:      render.BuildSettingsHubBreadcrumbs(ctx),
		ViewStylesheetURLs:   []string{"/static/css/sumeru-settings-hub.css"},
		ExtraStylesheetURLs:  render.ExtraStylesheetURLs,
		ExtraBodyClasses:     " sum-body--settings-hub",
	}

	html, err := render.RenderPage(ctx, config.AppConfig.TemplatesPath, page)
	if err != nil {
		http.Error(w, fmt.Sprintf("Layout: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
