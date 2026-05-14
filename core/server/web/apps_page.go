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

// appsModuleDetailVM is the readonly / edit form for one ir.module row on the Apps screen.
type appsModuleDetailVM struct {
	Layout                                    string
	Editing                                   bool
	Name, DisplayName, Author, Version        string
	Description, State                        string
	Active                                    bool
	ID                                        int
	CanInstall, CanUninstall                  bool
	CanDeactivate, CanActivate                bool
	BackAppsQuery                             string // layout=… (no module)
	EditURL, CancelURL                        string
}

type appsPageData struct {
	Title         string
	Message       string
	Modules       []appsModule
	Layout        string
	ModuleDetail  *appsModuleDetailVM
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

	var detail *appsModuleDetailVM
	breadcrumb := "Applications"
	if moduleParam != "" {
		row, err := orm.SearchOne(r.Context(), "ir.module", map[string]interface{}{"name": moduleParam})
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
		backQ := "layout=" + layout
		qEdit := url.Values{}
		qEdit.Set("module", name)
		qEdit.Set("layout", layout)
		qEdit.Set("edit", "1")
		qCancel := url.Values{}
		qCancel.Set("module", name)
		qCancel.Set("layout", layout)
		detail = &appsModuleDetailVM{
			Layout:          layout,
			Editing:         editing,
			Name:            name,
			DisplayName:     stringField(row["display_name"]),
			Author:          stringField(row["author"]),
			Version:         stringField(row["version"]),
			Description:     stringField(row["description"]),
			State:           stringField(row["state"]),
			Active:          boolField(row["active"]),
			ID:              int(id64),
			CanInstall:      found.CanInstall,
			CanUninstall:    found.CanUninstall,
			CanDeactivate:   found.CanDeactivate,
			CanActivate:     found.CanActivate,
			BackAppsQuery:   backQ,
			EditURL:         "/web/apps?" + qEdit.Encode(),
			CancelURL:       "/web/apps?" + qCancel.Encode(),
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
		Title:           "Apps",
		Message:         msg,
		Modules:         mods,
		Layout:          layout,
		ModuleDetail:    detail,
		ViewBreadcrumb:  breadcrumb,
	}
	if err := innerTmpl.Execute(&innerBuf, data); err != nil {
		log.Printf("%s: execute apps_inner: %v", platformmsg.MsgHTTPTemplateError, err)
		http.Error(w, fmt.Sprintf("Template execute: %v", err), http.StatusInternalServerError)
		return
	}

	topMenus, sidebarMenus, activeModuleID := render.LoadShellMenus(r.Context(), "")
	page := render.PageData{
		Title:               "Apps",
		ViewBreadcrumb:      breadcrumb,
		ModuleName:          "Apps",
		Content:             template.HTML(innerBuf.String()),
		TopMenus:            topMenus,
		SidebarMenus:        sidebarMenus,
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        "",
		ViewStylesheetURLs:  []string{"/static/css/view-apps.css"},
		AppsNavActive:       true,
		ExtraStylesheetURLs: render.ExtraStylesheetURLs,
		ViewTabs:            render.AppsViewTabs(layout, msg, moduleParam),
	}
	if detail != nil {
		page.ActivityContextModel = "ir.module"
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
