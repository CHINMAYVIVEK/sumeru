package web

import (
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/orm"
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
	CSRFToken      string
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
	if !requireSystemAdmin(w, r, true) {
		return
	}
	ctx := r.Context()
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

	raw, err := module.ListModules(ctx)
	if err != nil {
		WebLogEvent(ctx, r.URL.Path, "Failed to list modules for apps page", "load", "failure", err, nil)
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
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
		am.IconLetter = render.IconLetterFromName(am.DisplayName)
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
		ViewStylesheetURLs:  []string{"/static/css/sumeru-apps.css"},
		AppsNavActive:       true,
		ViewTabs:            render.AppsViewTabs(layout, msg, moduleParam, filter, scope, searchQ),
		BreadcrumbItems:     render.BuildAppsBreadcrumbs(ctx, listHref, detailTitle),
	}
	if detail != nil {
		page.ActivityContextModel = "sys.module"
		page.ActivityContextRecordID = int64(detail.ID)
	}

	renderShellPage(w, r, shellPageOpts{
		Route:               "/web/apps",
		InnerTemplate:       "apps_inner.html",
		InnerData: appsPageData{
			Title:          "Apps",
			Message:        msg,
			CSRFToken:      CSRFTokenForRequest(r),
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
		},
		Page:                page,
		ExtraStylesheetURLs: []string{"/static/css/sumeru-apps.css"},
	})

	navFields := map[string]interface{}{
		"layout": layout,
		"filter": filter,
		"scope":  scope,
		"search": searchQ,
	}
	if moduleParam != "" {
		navFields["module"] = moduleParam
	}
	WebLogNavigation(ctx, r.URL.Path, "apps_open", "Apps page opened", navFields)
}
