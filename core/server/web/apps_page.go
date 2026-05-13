package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"sumeru/core/base/platformmsg"
	"sumeru/core/engine/render"
	"sumeru/core/module"
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

type appsKanbanColumn struct {
	Title    string
	Subtitle string
	Modules  []appsModule
}

type appsPageData struct {
	Title      string
	Message    string
	Modules    []appsModule
	Layout     string
	KanbanCols []appsKanbanColumn
}

// AppsHandler lists installable apps and exposes install / uninstall / activate controls.
func AppsHandler(w http.ResponseWriter, r *http.Request) {
	msg := strings.TrimSpace(r.URL.Query().Get("msg"))
	layout := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("layout")))
	if layout == "" {
		layout = "grid"
	}
	if layout != "grid" && layout != "kanban" && layout != "list" {
		layout = "grid"
	}

	raw, err := module.ListModules()
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

	var discover, activeOn, activeOff []appsModule
	for _, m := range mods {
		if !m.Application {
			continue
		}
		if m.State != "installed" {
			discover = append(discover, m)
		} else if m.Active {
			activeOn = append(activeOn, m)
		} else {
			activeOff = append(activeOff, m)
		}
	}
	kanbanCols := []appsKanbanColumn{
		{Title: "Discover", Subtitle: "Ready to activate", Modules: discover},
		{Title: "Running", Subtitle: "Installed and on", Modules: activeOn},
		{Title: "Paused", Subtitle: "Installed but disabled", Modules: activeOff},
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
		Title:      "Apps",
		Message:    msg,
		Modules:    mods,
		Layout:     layout,
		KanbanCols: kanbanCols,
	}
	if err := innerTmpl.Execute(&innerBuf, data); err != nil {
		log.Printf("%s: execute apps_inner: %v", platformmsg.MsgHTTPTemplateError, err)
		http.Error(w, fmt.Sprintf("Template execute: %v", err), http.StatusInternalServerError)
		return
	}

	topMenus, sidebarMenus, activeModuleID := render.LoadShellMenus("")
	page := render.PageData{
		Title:               "Apps",
		ViewBreadcrumb:      "Applications",
		ModuleName:          "Apps",
		Content:             template.HTML(innerBuf.String()),
		TopMenus:            topMenus,
		SidebarMenus:        sidebarMenus,
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        "",
		ViewStylesheetURLs:  []string{"/static/css/view-apps.css"},
		AppsNavActive:       true,
		ExtraStylesheetURLs: render.ExtraStylesheetURLs,
		ViewTabs:            render.AppsViewTabs(layout, msg),
	}
	html, err := render.RenderPage(config.AppConfig.TemplatesPath, page)
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
