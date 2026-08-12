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

type appsModule struct {
	Name          string
	DisplayName   string
	Author        string
	Version       string
	Description   string
	State         string
	Application   bool
	Active        bool
	IsCore        bool
	CanInstall    bool
	CanUninstall  bool
	CanDeactivate bool
	CanActivate   bool
	IconLetter    string // first letter for app tile
}

type appsPageData struct {
	Title          string
	Message        string
	Modules        []appsModule
	AppModules     []appsModule
	TechModules    []appsModule
	Layout         string
	Filter         string
	Scope          string
	Search         string
	Nav            appsNavVM
	ModuleDetail   *appsModuleDetailVM
	ViewBreadcrumb string
}

// AppsHandler lists installable apps and exposes install / uninstall / activate controls.
func AppsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !orm.UserHasGroupXML(r.Context(), orm.SecurityUID(r.Context()), "base.group_system") {
		http.Redirect(w, r, "/web/home", http.StatusFound)
		return
	}
	msg := strings.TrimSpace(r.URL.Query().Get("msg"))
	layout := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("layout")))
	if layout == "" {
		layout = "grid"
	}
	if layout == "kanban" {
		layout = "grid"
	}
	if layout != "grid" && layout != "list" {
		layout = "grid"
	}

	moduleParam := strings.TrimSpace(r.URL.Query().Get("module"))
	editing := strings.TrimSpace(r.URL.Query().Get("edit")) == "1"
	filter := normalizeAppsFilter(r.URL.Query().Get("filter"))
	scope := normalizeAppsScope(r.URL.Query().Get("scope"))
	searchQ := strings.TrimSpace(r.URL.Query().Get("q"))

	raw, err := module.ListModules(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list modules: %v", err), http.StatusInternalServerError)
		return
	}

	var mods []appsModule
	for _, row := range raw {
		name := orm.AsString(row["name"])
		if name == "" {
			continue
		}
		state := orm.AsString(row["state"])
		active := orm.AsBool(row["active"])
		app := orm.AsBool(row["application"])

		am := appsModule{
			Name:          name,
			DisplayName:   orm.AsString(row["display_name"]),
			Author:        orm.AsString(row["author"]),
			Version:       orm.AsString(row["version"]),
			Description:   orm.AsString(row["description"]),
			State:         state,
			Application:   app,
			Active:        active,
			IsCore:        name == "base",
			CanInstall:    state != "installed",
			CanUninstall:  state == "installed" && name != "base",
			CanDeactivate: state == "installed" && active && name != "base",
			CanActivate:   state == "installed" && !active && name != "base",
		}
		if am.DisplayName == "" {
			am.DisplayName = am.Name
		}
		am.IconLetter = iconLetterFromName(am.DisplayName)
		mods = append(mods, am)
	}

	var appMods, techMods []appsModule
	for _, m := range mods {
		if !appsModuleMatchesSearch(m, searchQ) || !appsModuleMatchesFilter(m, filter) {
			continue
		}
		if m.Application {
			appMods = append(appMods, m)
		} else {
			techMods = append(techMods, m)
		}
	}
	switch scope {
	case "apps":
		techMods = nil
	case "technical":
		appMods = nil
	}

	detail, breadcrumb, ok := loadAppsModuleDetail(w, r, moduleParam, editing, mods, layout, filter, scope, searchQ)
	if !ok {
		return
	}

	tmplPath := filepath.Join(config.AppConfig.TemplatesPath, "apps_inner.html")
	innerTmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		log.Printf("%s: parse %s: %v", platformmsg.MsgHTTPTemplateError, tmplPath, err)
		http.Error(w, platformmsg.MsgHTTPTemplateError, http.StatusInternalServerError)
		return
	}
	var innerBuf bytes.Buffer
	data := appsPageData{
		Title:          "Apps",
		Message:        msg,
		Modules:        mods,
		AppModules:     appMods,
		TechModules:    techMods,
		Layout:         layout,
		Filter:         filter,
		Scope:          scope,
		Search:         searchQ,
		Nav:            buildAppsNavVM(layout, filter, scope, searchQ),
		ModuleDetail:   detail,
		ViewBreadcrumb: breadcrumb,
	}
	if err := innerTmpl.Execute(&innerBuf, data); err != nil {
		log.Printf("%s: execute apps_inner: %v", platformmsg.MsgHTTPTemplateError, err)
		http.Error(w, fmt.Sprintf("Template execute: %v", err), http.StatusInternalServerError)
		return
	}

	topMenus, sidebarMenus, activeModuleID, _ := render.LoadShellMenus(r.Context(), "")
	listHref := "/web/apps"
	if bq := appsBrowseQuery(layout, filter, scope, searchQ); bq != "" {
		listHref = "/web/apps?" + bq
	}
	detailTitle := ""
	if detail != nil {
		detailTitle = detail.DisplayName
	}
	page := render.PageData{
		Title:               "Apps",
		ViewBreadcrumb:      breadcrumb,
		ModuleName:          "Apps",
		Content:             template.HTML(innerBuf.String()),
		TopMenus:            topMenus,
		SidebarMenus:        sidebarMenus,
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        "",
		ViewStylesheetURLs:  []string{"/static/css/sumeru-apps.css"},
		AppsNavActive:       true,
		ExtraStylesheetURLs: render.ExtraStylesheetURLs,
		ViewTabs:            render.AppsViewTabs(layout, msg, moduleParam, filter, scope, searchQ),
		BreadcrumbItems:     render.BuildAppsBreadcrumbs(r.Context(), listHref, detailTitle),
	}
	if detail != nil {
		page.ActivityContextModel = "sys.module"
		page.ActivityContextRecordID = int64(detail.ID)
	}
	html, err := render.RenderPage(r.Context(), config.AppConfig.TemplatesPath, page)
	if err != nil {
		http.Error(w, fmt.Sprintf("Layout render: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}
