package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"sumeru/core/base/platformmsg"
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

// appsModuleDetailVM is the readonly / edit form for one sys.module row on the Apps screen.
type appsModuleDetailVM struct {
	Layout                             string
	Editing                            bool
	Name, DisplayName, Author, Version string
	Description, State                 string
	Active                             bool
	ID                                 int
	CanInstall, CanUninstall           bool
	CanDeactivate, CanActivate         bool
	BackAppsQuery                      string // query string without leading ?
	EditURL, CancelURL                 string
}

type appsNavVM struct {
	FilterAll, FilterInstalled, FilterUninstalled string
	ScopeAll, ScopeApps, ScopeTechnical           string
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
		name := stringField(row["name"])
		if name == "" {
			continue
		}
		state := stringField(row["state"])
		active := boolField(row["active"])
		app := boolField(row["application"])

		am := appsModule{
			Name:          name,
			DisplayName:   stringField(row["display_name"]),
			Author:        stringField(row["author"]),
			Version:       stringField(row["version"]),
			Description:   stringField(row["description"]),
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
		if r := []rune(strings.TrimSpace(am.DisplayName)); len(r) > 0 {
			am.IconLetter = strings.ToUpper(string(r[0]))
		} else {
			am.IconLetter = "?"
		}
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

	var detail *appsModuleDetailVM
	breadcrumb := "Applications"
	if moduleParam != "" {
		row, err := orm.SearchOne(r.Context(), "sys.module", map[string]interface{}{"name": moduleParam})
		if err != nil {
			http.Error(w, "Module not found", http.StatusNotFound)
			return
		}
		id64, ok := orm.CoerceInt64(row["id"])
		if !ok {
			http.Error(w, "Module not found", http.StatusNotFound)
			return
		}
		name := stringField(row["name"])
		var found appsModule
		for _, m := range mods {
			if m.Name == name {
				found = m
				break
			}
		}
		backQ := appsBrowseQuery(layout, filter, scope, searchQ)
		qEdit := url.Values{}
		appendAppsQueryBase(qEdit, layout, filter, scope, searchQ)
		qEdit.Set("module", name)
		qEdit.Set("edit", "1")
		qCancel := url.Values{}
		appendAppsQueryBase(qCancel, layout, filter, scope, searchQ)
		qCancel.Set("module", name)
		detail = &appsModuleDetailVM{
			Layout:        layout,
			Editing:       editing,
			Name:          name,
			DisplayName:   stringField(row["display_name"]),
			Author:        stringField(row["author"]),
			Version:       stringField(row["version"]),
			Description:   stringField(row["description"]),
			State:         stringField(row["state"]),
			Active:        boolField(row["active"]),
			ID:            int(id64),
			CanInstall:    found.CanInstall,
			CanUninstall:  found.CanUninstall,
			CanDeactivate: found.CanDeactivate,
			CanActivate:   found.CanActivate,
			BackAppsQuery: backQ,
			EditURL:       "/web/apps?" + qEdit.Encode(),
			CancelURL:     "/web/apps?" + qCancel.Encode(),
		}
		if detail.DisplayName == "" {
			detail.DisplayName = detail.Name
		}
		breadcrumb = "Apps · " + detail.DisplayName
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

func stringField(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

func boolField(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case []byte:
		return string(t) == "t" || string(t) == "true" || string(t) == "1"
	default:
		s := strings.ToLower(stringField(v))
		return s == "true" || s == "t" || s == "1"
	}
}

func normalizeAppsFilter(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "installed", "uninstalled":
		return strings.ToLower(strings.TrimSpace(s))
	default:
		return "all"
	}
}

func normalizeAppsScope(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "apps", "technical":
		return strings.ToLower(strings.TrimSpace(s))
	default:
		return "all"
	}
}

func appsModuleMatchesFilter(m appsModule, filter string) bool {
	switch filter {
	case "installed":
		return m.State == "installed"
	case "uninstalled":
		return m.State != "installed"
	default:
		return true
	}
}

func appsModuleMatchesSearch(m appsModule, q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	}
	lq := strings.ToLower(q)
	return strings.Contains(strings.ToLower(m.Name), lq) || strings.Contains(strings.ToLower(m.DisplayName), lq)
}

func appendAppsQueryBase(v url.Values, layout, filter, scope, q string) {
	if layout == "list" || layout == "grid" {
		v.Set("layout", layout)
	}
	if filter != "" && filter != "all" {
		v.Set("filter", filter)
	}
	if scope != "" && scope != "all" {
		v.Set("scope", scope)
	}
	if strings.TrimSpace(q) != "" {
		v.Set("q", strings.TrimSpace(q))
	}
}

func appsBrowseQuery(layout, filter, scope, q string) string {
	v := url.Values{}
	appendAppsQueryBase(v, layout, filter, scope, q)
	return v.Encode()
}

func appsLink(layout, filter, scope, q string) string {
	v := url.Values{}
	appendAppsQueryBase(v, layout, filter, scope, q)
	return "/web/apps?" + v.Encode()
}

func buildAppsNavVM(layout, filter, scope, q string) appsNavVM {
	return appsNavVM{
		FilterAll:         appsLink(layout, "all", scope, q),
		FilterInstalled:   appsLink(layout, "installed", scope, q),
		FilterUninstalled: appsLink(layout, "uninstalled", scope, q),
		ScopeAll:          appsLink(layout, filter, "all", q),
		ScopeApps:         appsLink(layout, filter, "apps", q),
		ScopeTechnical:    appsLink(layout, filter, "technical", q),
	}
}
