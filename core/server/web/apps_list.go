package web

import (
	"context"
	"net/http"

	"sumeru/core/engine/render"
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
	browse := parseAppsBrowseState(r)

	moduleRows, ok := listModulesOr500(w, r, "Failed to list modules for apps page")
	if !ok {
		return
	}

	allModules := buildAppsModuleList(moduleRows)
	appModules, techModules := filterAppsModulesByBrowse(allModules, browse)

	detail, breadcrumb, ok := loadAppsModuleDetail(
		w, r,
		browse.ModuleName,
		browse.Editing,
		allModules,
		browse.Layout,
		browse.Filter,
		browse.Scope,
		browse.SearchQuery,
	)
	if !ok {
		return
	}

	listHref := appsLinkFromBrowse(browse)
	detailTitle := ""
	if detail != nil {
		detailTitle = detail.DisplayName
	}

	page := render.PageData{
		Title:              appsPageTitle,
		ViewBreadcrumb:     breadcrumb,
		ModuleName:         appsPageTitle,
		ViewStylesheetURLs: []string{appsStylesheetURL},
		AppsNavActive:      true,
		SuppressSidebar:    true,
		ViewTabs: render.AppsViewTabs(
			browse.Layout,
			browse.Message,
			browse.ModuleName,
			browse.Filter,
			browse.Scope,
			browse.SearchQuery,
		),
		BreadcrumbItems: render.BuildAppsBreadcrumbs(ctx, listHref, detailTitle),
	}
	if detail != nil {
		page.ActivityContextModel = appsModuleModel
		page.ActivityContextRecordID = int64(detail.ID)
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         appsRoute,
		InnerTemplate: appsInnerTemplate,
		InnerData: appsPageData{
			Title:          appsPageTitle,
			Message:        browse.Message,
			CSRFToken:      CSRFTokenForRequest(r),
			Modules:        allModules,
			AppModules:     appModules,
			TechModules:    techModules,
			Layout:         browse.Layout,
			Filter:         browse.Filter,
			Scope:          browse.Scope,
			Search:         browse.SearchQuery,
			Nav:            buildAppsNavVM(browse),
			ModuleDetail:   detail,
			ViewBreadcrumb: breadcrumb,
		},
		Page: page,
	})

	logAppsPageOpen(ctx, r.URL.Path, browse)
}

func buildAppsModuleList(moduleRows []map[string]interface{}) []appsModule {
	modules := make([]appsModule, 0, len(moduleRows))
	for _, row := range moduleRows {
		parsed, rowOK := parseModuleRow(row)
		if !rowOK {
			continue
		}
		modules = append(modules, appsModuleFromParsed(parsed))
	}
	return modules
}

// appsModuleFromParsed maps a normalized module row to the Apps list view model, including action flags.
func appsModuleFromParsed(parsed moduleRow) appsModule {
	isCore := parsed.Name == "base"
	isInstalled := parsed.State == "installed"
	return appsModule{
		Name:          parsed.Name,
		DisplayName:   parsed.DisplayName,
		Author:        parsed.Author,
		Version:       parsed.Version,
		Description:   parsed.Description,
		State:         parsed.State,
		Application:   parsed.Application,
		Active:        parsed.Active,
		IsCore:        isCore,
		CanInstall:    !isInstalled,
		CanUninstall:  isInstalled && !isCore,
		CanDeactivate: isInstalled && parsed.Active && !isCore,
		CanActivate:   isInstalled && !parsed.Active && !isCore,
		IconLetter:    render.IconLetterFromName(parsed.DisplayName),
	}
}

func logAppsPageOpen(ctx context.Context, route string, browse appsBrowseState) {
	fields := map[string]interface{}{
		"layout": browse.Layout,
		"filter": browse.Filter,
		"scope":  browse.Scope,
		"search": browse.SearchQuery,
	}
	if browse.ModuleName != "" {
		fields["module"] = browse.ModuleName
	}
	WebLogNavigation(ctx, route, "apps_open", "Apps page opened", fields)
}
