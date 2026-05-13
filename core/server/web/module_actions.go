package web

import (
	"net/http"
	"net/url"
	"strings"

	"sumeru/core/module"
)

// ModuleActionHandler handles POST actions: install, uninstall, activate, deactivate.
func ModuleActionHandler(w http.ResponseWriter, r *http.Request) {
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
		err = module.InstallModuleByName(mod)
		msg = "installed_" + mod
	case "uninstall":
		err = module.UninstallModuleByName(mod)
		msg = "uninstalled_" + mod
	case "deactivate":
		err = module.SetModuleActive(mod, false)
		msg = "deactivated_" + mod
	case "activate":
		err = module.SetModuleActive(mod, true)
		msg = "activated_" + mod
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
	if layout == "kanban" || layout == "list" || layout == "grid" {
		q.Set("layout", layout)
	}
	if enc := q.Encode(); enc != "" {
		return "/web/apps?" + enc
	}
	return "/web/apps"
}
