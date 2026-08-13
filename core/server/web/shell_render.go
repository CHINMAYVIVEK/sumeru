package web

import (
	"bytes"
	"context"
	"html/template"
	"net/http"
	"path/filepath"

	"sumeru/core/engine/render"
	"sumeru/core/server/config"
)

type shellPageOpts struct {
	Route               string
	InnerTemplate       string
	InnerData           interface{}
	MenuIDStr           string
	Page                render.PageData
	ExtraStylesheetURLs []string
}

func renderShellPage(w http.ResponseWriter, r *http.Request, opts shellPageOpts) {
	ctx := r.Context()
	route := opts.Route
	if route == "" {
		route = r.URL.Path
	}
	innerPath := filepath.Join(config.AppConfig.TemplatesPath, opts.InnerTemplate)
	tmpl, err := template.ParseFiles(innerPath)
	if err != nil {
		WebLogEvent(ctx, route, "Failed to parse inner template", "render", "failure", err,
			map[string]interface{}{"template": opts.InnerTemplate})
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}
	var innerBuf bytes.Buffer
	if err := tmpl.Execute(&innerBuf, opts.InnerData); err != nil {
		WebLogEvent(ctx, route, "Failed to execute inner template", "render", "failure", err, nil)
		http.Error(w, "Template error", http.StatusInternalServerError)
		return
	}

	topMenus, sidebarMenus, activeModuleID, moduleName := render.LoadShellMenus(ctx, opts.MenuIDStr)
	page := opts.Page
	if page.Title == "" {
		page.Title = "Sumeru"
	}
	page.Content = template.HTML(innerBuf.String())
	page.TopMenus = topMenus
	page.SidebarMenus = sidebarMenus
	page.ActiveModuleID = activeModuleID
	if page.ModuleName == "" {
		page.ModuleName = moduleName
	}
	if page.ActiveMenuID == "" {
		page.ActiveMenuID = opts.MenuIDStr
	}
	if len(page.ViewStylesheetURLs) == 0 {
		page.ViewStylesheetURLs = []string{"/static/css/sumeru-workspace.css"}
	}
	if len(page.ExtraStylesheetURLs) == 0 {
		page.ExtraStylesheetURLs = render.ExtraStylesheetURLs
	}
	if len(opts.ExtraStylesheetURLs) > 0 {
		page.ExtraStylesheetURLs = opts.ExtraStylesheetURLs
	}
	if page.CSRFToken == "" {
		page.CSRFToken = CSRFTokenForRequest(r)
	}

	html, err := render.RenderPage(ctx, config.AppConfig.TemplatesPath, page)
	if err != nil {
		WebLogEvent(ctx, route, "Failed to render page layout", "render", "failure", err, nil)
		http.Error(w, "Layout render error", http.StatusInternalServerError)
		return
	}
	writeHTML(w, ctx, route, html)
}

func writeHTML(w http.ResponseWriter, ctx context.Context, route, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write([]byte(html)); err != nil {
		WebLogEvent(ctx, route, "Failed to write HTML response", "write", "partial", err, nil)
	}
}
