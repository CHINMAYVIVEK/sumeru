package web

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// AppLogsHandler serves Event Log inside the main shell (Settings sidebar, Companies / Users / App Logs).
func AppLogsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	reqCtx := r.Context()
	ctxBypass := orm.ContextWithBypass(reqCtx, true)

	events, err := loadAppLogEvents(ctxBypass)
	if err != nil {
		log.Printf("App Logs: query failed: %v", err)
		http.Error(w, "Failed to load app logs", http.StatusInternalServerError)
		return
	}

	innerPath := filepath.Join(config.AppConfig.TemplatesPath, "app_logs_inner.html")
	tmpl, err := template.ParseFiles(innerPath)
	if err != nil {
		log.Printf("App Logs: parse %s: %v", innerPath, err)
		http.Error(w, "App logs template missing", http.StatusInternalServerError)
		return
	}
	var innerBuf bytes.Buffer
	if err := tmpl.Execute(&innerBuf, events); err != nil {
		log.Printf("App Logs: inner template: %v", err)
		http.Error(w, "App logs render error", http.StatusInternalServerError)
		return
	}

	menuIDStr := ""
	if mid, _, err := orm.ResolveXmlId(ctxBypass, "base.menu_app_logs"); err == nil && mid > 0 {
		menuIDStr = fmt.Sprintf("%d", mid)
	}

	topMenus, sidebarMenus, activeModuleID, moduleName := render.LoadShellMenus(reqCtx, menuIDStr)
	page := render.PageData{
		Title:               "App Logs",
		ViewBreadcrumb:      "Event Log",
		ModuleName:          moduleName,
		Content:             template.HTML(innerBuf.String()),
		TopMenus:            topMenus,
		SidebarMenus:        sidebarMenus,
		ActiveModuleID:      activeModuleID,
		ActiveMenuID:        menuIDStr,
		ViewStylesheetURLs:  []string{"/static/css/sumeru-workspace.css"},
		ExtraStylesheetURLs: render.ExtraStylesheetURLs,
		SettingsNavActive:   true,
	}
	if mid, _, err := orm.ResolveXmlId(ctxBypass, "base.menu_app_logs"); err == nil && mid > 0 {
		page.BreadcrumbItems = render.BuildAppLogsBreadcrumbs(reqCtx, mid)
	}

	html, err := render.RenderPage(reqCtx, config.AppConfig.TemplatesPath, page)
	if err != nil {
		log.Printf("App Logs: layout: %v", err)
		http.Error(w, "Layout render error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

// ModuleEvent is one row for the app logs table template.
type ModuleEvent struct {
	CreatedAt string
	Module    string
	Action    string
	Detail    string
}

func loadAppLogEvents(ctx context.Context) ([]ModuleEvent, error) {
	tbl := orm.GetTableName("app.log")
	query := `
		SELECT
			to_char(create_date, 'YYYY-MM-DD HH24:MI:SS') AS created_at,
			COALESCE(NULLIF(TRIM(module_name), ''), '') AS module,
			COALESCE(NULLIF(TRIM(action), ''), '') AS action,
			COALESCE(NULLIF(TRIM(detail), ''), '') AS detail
		FROM ` + tbl + `
		ORDER BY create_date DESC, id DESC
		LIMIT 500`

	rows, err := orm.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ModuleEvent
	for rows.Next() {
		var e ModuleEvent
		if err := rows.Scan(&e.CreatedAt, &e.Module, &e.Action, &e.Detail); err != nil {
			continue
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
