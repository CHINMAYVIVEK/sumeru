package web

import (
	"context"
	"fmt"
	"net/http"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

// AppLogsHandler serves Event Log inside the main shell.
func AppLogsHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !requireMenuAccess(w, r, "base.menu_app_logs") {
		return
	}
	reqCtx := r.Context()

	events, err := loadAppLogEvents(reqCtx)
	if err != nil {
		WebLogEvent(reqCtx, r.URL.Path, "Failed to load app logs", "load", "failure", err, nil)
		http.Error(w, "Failed to load app logs", http.StatusInternalServerError)
		return
	}

	menuIDStr := ""
	var appLogsMenuID int
	if mid, _, err := orm.ResolveXmlId(reqCtx, "base.menu_app_logs"); err == nil && mid > 0 {
		appLogsMenuID = mid
		menuIDStr = fmt.Sprintf("%d", mid)
	}

	page := render.PageData{
		Title:              "App Logs",
		ViewBreadcrumb:     "Event Log",
		SettingsNavActive:  true,
		ViewStylesheetURLs: []string{
			"/static/css/sumeru-workspace.css",
			"/static/css/sumeru-pages.css",
		},
	}
	if appLogsMenuID > 0 {
		page.BreadcrumbItems = render.BuildAppLogsBreadcrumbs(reqCtx, appLogsMenuID)
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         r.URL.Path,
		InnerTemplate: "app_logs_inner.html",
		InnerData:     events,
		MenuIDStr:     menuIDStr,
		Page:          page,
	})
}

type ModuleEvent struct {
	CreatedAt string
	Module    string
	Action    string
	Detail    string
}

func loadAppLogEvents(ctx context.Context) ([]ModuleEvent, error) {
	tbl := orm.MustQuotedTableName("app.log")
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
			WebLogEvent(ctx, "-", "App log row scan failed", "load", "partial", err, nil)
			continue
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
