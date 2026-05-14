package web

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// ModuleActionHandler handles POST actions: install, uninstall, activate, deactivate.
func ModuleActionHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), "ir.module", "write"); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/web/apps", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, appsRedirectPath("invalid_form", r.FormValue("apps_layout")), http.StatusSeeOther)
		return
	}
	action := strings.TrimSpace(r.FormValue("do"))
	mod := strings.TrimSpace(r.FormValue("module"))
	if mod == "" {
		http.Redirect(w, r, appsRedirectPath("missing_module", r.FormValue("apps_layout")), http.StatusSeeOther)
		return
	}

	var err error
	var msg string
	switch action {
	case "install":
		err = module.InstallModuleByName(r.Context(), mod)
		msg = "installed_" + mod
	case "uninstall":
		err = module.UninstallModuleByName(r.Context(), mod)
		msg = "uninstalled_" + mod
	case "deactivate":
		err = module.SetModuleActive(r.Context(), mod, false)
		msg = "deactivated_" + mod
	case "activate":
		err = module.SetModuleActive(r.Context(), mod, true)
		msg = "activated_" + mod
	case "save_module":
		err = saveModuleFromForm(r)
		if err != nil {
			http.Redirect(w, r, appsRedirectPath(err.Error(), r.FormValue("apps_layout")), http.StatusSeeOther)
			return
		}
		q := url.Values{}
		q.Set("msg", "saved")
		if l := strings.ToLower(strings.TrimSpace(r.FormValue("apps_layout"))); l == "list" || l == "grid" {
			q.Set("layout", l)
		}
		q.Set("module", mod)
		http.Redirect(w, r, "/web/apps?"+q.Encode(), http.StatusSeeOther)
		return
	default:
		http.Redirect(w, r, appsRedirectPath("unknown_action", r.FormValue("apps_layout")), http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Redirect(w, r, appsRedirectPath(err.Error(), r.FormValue("apps_layout")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, appsRedirectPath(msg, r.FormValue("apps_layout")), http.StatusSeeOther)
}

func appsRedirectPath(msg, layout string) string {
	q := url.Values{}
	if strings.TrimSpace(msg) != "" {
		q.Set("msg", strings.TrimSpace(msg))
	}
	layout = strings.ToLower(strings.TrimSpace(layout))
	if layout == "kanban" {
		layout = "grid"
	}
	if layout == "list" || layout == "grid" {
		q.Set("layout", layout)
	}
	if enc := q.Encode(); enc != "" {
		return "/web/apps?" + enc
	}
	return "/web/apps"
}

func saveModuleFromForm(r *http.Request) error {
	mod := strings.TrimSpace(r.FormValue("module"))
	if mod == "" {
		return fmt.Errorf("missing_module")
	}
	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("module_row_id")))
	if err != nil || id <= 0 {
		return fmt.Errorf("invalid_module_row")
	}
	row, err := orm.SearchOne(r.Context(), "ir.module", map[string]interface{}{"id": id})
	if err != nil {
		return fmt.Errorf("module_not_found")
	}
	if strings.TrimSpace(orm.AsString(row["name"])) != mod {
		return fmt.Errorf("module_mismatch")
	}
	vals := map[string]interface{}{
		"display_name": strings.TrimSpace(r.FormValue("display_name")),
		"author":       strings.TrimSpace(r.FormValue("author")),
		"description":  strings.TrimSpace(r.FormValue("description")),
	}
	return orm.UpdateRecordByID(r.Context(), "ir.module", id, vals)
}
