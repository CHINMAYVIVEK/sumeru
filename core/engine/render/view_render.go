package render

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/addons/mail"
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
	}
	if strings.EqualFold(view.Type, "form") && view.Chatter != nil && recData.RecordID > 0 &&
		mail.CompanyChatterEnabled(ctx) && mail.CompanyActivityPanelEnabled(ctx) {
		pageData.ActivityPanelChatter = true
		var ch strings.Builder
		writeActivityChatterPanel(ctx, &ch, view.Chatter, recData, view.Model)
		pageData.ActivityChatterHTML = template.HTML(ch.String())
	}

	log.Printf("Rendering view for model %s (ActiveMenu: %s, ActiveModule: %s)", view.Model, activeMenuID, activeModuleID)

	out, err := RenderPage(ctx, templatesDir, pageData)
	if err != nil {
		log.Printf("Error rendering page: %v", err)
		return content
	}
	return out
}
