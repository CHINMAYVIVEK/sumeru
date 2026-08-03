package web

import (
	"net/http"
	"net/url"

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

// loadAppsModuleDetail builds the detail VM for a module query param, or nil when moduleParam is empty.
// On DB/lookup failure it writes an HTTP error and returns ok=false.
func loadAppsModuleDetail(w http.ResponseWriter, r *http.Request, moduleParam string, editing bool, mods []appsModule, layout, filter, scope, searchQ string) (detail *appsModuleDetailVM, breadcrumb string, ok bool) {
	breadcrumb = "Applications"
	if moduleParam == "" {
		return nil, breadcrumb, true
	}

	row, err := orm.SearchOne(r.Context(), "sys.module", map[string]interface{}{"name": moduleParam})
	if err != nil {
		http.Error(w, "Module not found", http.StatusNotFound)
		return nil, "", false
	}
	id64, idOK := orm.CoerceInt64(row["id"])
	if !idOK {
		http.Error(w, "Module not found", http.StatusNotFound)
		return nil, "", false
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
	return detail, breadcrumb, true
}
