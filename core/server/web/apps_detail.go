package web

import (
	"net/http"

	"sumeru/core/orm"
)

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

const appsListBreadcrumb = "Applications"

// loadAppsModuleDetail builds the detail VM when ?module= is present, or nil for list-only view.
// On lookup failure it writes an HTTP error and returns ok=false.
func loadAppsModuleDetail(
	w http.ResponseWriter,
	r *http.Request,
	moduleName string,
	editing bool,
	listedModules []appsModule,
	layout, filter, scope, searchQuery string,
) (detail *appsModuleDetailVM, breadcrumb string, ok bool) {
	breadcrumb = appsListBreadcrumb
	if moduleName == "" {
		return nil, breadcrumb, true
	}

	parsed, recordID, loaded := loadModuleRowByName(w, r, moduleName)
	if !loaded {
		return nil, "", false
	}

	// Action flags come from the pre-built list entry when available (same rules as the grid/list).
	listEntry, _ := findAppsModule(listedModules, parsed.Name)
	browseQuery := appsBrowseQuery(layout, filter, scope, searchQuery)

	detail = &appsModuleDetailVM{
		Layout:        layout,
		Editing:       editing,
		Name:          parsed.Name,
		DisplayName:   parsed.DisplayName,
		Author:        parsed.Author,
		Version:       parsed.Version,
		Description:   parsed.Description,
		State:         parsed.State,
		Active:        parsed.Active,
		ID:            recordID,
		CanInstall:    listEntry.CanInstall,
		CanUninstall:  listEntry.CanUninstall,
		CanDeactivate: listEntry.CanDeactivate,
		CanActivate:   listEntry.CanActivate,
		BackAppsQuery: browseQuery,
		EditURL:       appsDetailURL(layout, filter, scope, searchQuery, parsed.Name, true),
		CancelURL:     appsDetailURL(layout, filter, scope, searchQuery, parsed.Name, false),
	}
	breadcrumb = "Apps · " + detail.DisplayName
	return detail, breadcrumb, true
}

// loadModuleRowByName fetches one sys.module row by technical name.
func loadModuleRowByName(w http.ResponseWriter, r *http.Request, moduleName string) (moduleRow, int, bool) {
	row, err := orm.SearchOne(r.Context(), appsModuleModel, map[string]interface{}{"name": moduleName})
	if err != nil {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	parsed, rowOK := parseModuleRow(row)
	if !rowOK {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	recordID, idOK := orm.CoerceInt64(row["id"])
	if !idOK {
		http.Error(w, "Module not found", http.StatusNotFound)
		return moduleRow{}, 0, false
	}
	return parsed, int(recordID), true
}

func findAppsModule(modules []appsModule, name string) (appsModule, bool) {
	for _, moduleEntry := range modules {
		if moduleEntry.Name == name {
			return moduleEntry, true
		}
	}
	return appsModule{}, false
}
