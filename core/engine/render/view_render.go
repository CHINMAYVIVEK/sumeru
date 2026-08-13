package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"sumeru/addons/mail"
	"sumeru/core/applog"
	"sumeru/core/engine/parser"
)

// RenderView builds full HTML for a workspace view (form, tree, kanban, pivot) inside the shell layout.
func RenderView(ctx context.Context, view *parser.View, activeMenuID, templatesDir string, recData *ViewRecordData) string {
	if recData == nil {
		recData = &ViewRecordData{}
	}
	var content string
	switch view.Type {
	case "form":
		content = RenderForm(ctx, view, recData)
	case "tree", "list":
		content = RenderTree(ctx, view, recData.ListRows, recData.ActionID, activeMenuID)
	case "kanban":
		content = RenderKanban(ctx, view, recData.ListRows, recData.ActionID, activeMenuID)
	case "pivot":
		content = RenderPivot(ctx, view)
	default:
		content = RenderForm(ctx, view, recData)
	}

	topMenus, sidebarMenus, activeModuleID, moduleName := LoadShellMenus(ctx, activeMenuID)
	viewBC := HumanViewBreadcrumb(view.Model, view.Type)

	actCtxModel := ""
	var actCtxID int64
	if strings.EqualFold(view.Type, "form") && recData.RecordID > 0 {
		actCtxModel = strings.TrimSpace(view.Model)
		actCtxID = int64(recData.RecordID)
	}

	pageData := PageData{
		Title:                   fmt.Sprintf("%s · %s", view.Model, view.Type),
		ViewBreadcrumb:          viewBC,
		ModuleName:              moduleName,
		Content:                 template.HTML(content),
		TopMenus:                topMenus,
		SidebarMenus:            sidebarMenus,
		ActiveModuleID:          activeModuleID,
		ActiveMenuID:            activeMenuID,
		ViewStylesheetURLs:      []string{"/static/css/sumeru-workspace.css"},
		ExtraStylesheetURLs:     ExtraStylesheetURLs,
		ViewTabs:                recData.ViewTabs,
		ActivityContextModel:    actCtxModel,
		ActivityContextRecordID: actCtxID,
		SettingsNavActive:       IsMenuUnderSettingsRoot(ctx, activeMenuID),
		BreadcrumbItems:         BuildWorkspaceBreadcrumbs(ctx, activeMenuID, view.Type, viewBC, recData.FormBaseQuery, recData.Record, recData.RecordID),
		CSRFToken:               recData.CSRFToken,
		FlashMessages:           recData.FlashMessages,
	}
	if strings.EqualFold(view.Type, "form") && view.Chatter != nil && recData.RecordID > 0 &&
		mail.CompanyChatterEnabled(ctx) && mail.CompanyActivityPanelEnabled(ctx) {
		pageData.ActivityPanelChatter = true
		var ch strings.Builder
		writeActivityChatterPanel(ctx, &ch, view.Chatter, recData, view.Model)
		pageData.ActivityChatterHTML = template.HTML(ch.String())
	}

	applog.DebugMsg(ctx, "render", "view",
		fmt.Sprintf("Rendering view for model %s", view.Model),
		map[string]interface{}{"active_menu": activeMenuID, "active_module": activeModuleID})

	out, err := RenderPage(ctx, templatesDir, pageData)
	if err != nil {
		applog.WarnMsg(ctx, "render", "view", "Error rendering page", err,
			map[string]interface{}{"model": view.Model})
		return content
	}
	return out
}
