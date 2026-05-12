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
		http.Redirect(w, r, "/web/apps?msg=invalid_form", http.StatusSeeOther)
		return
	}
	action := strings.TrimSpace(r.FormValue("do"))
	mod := strings.TrimSpace(r.FormValue("module"))
	if mod == "" {
		http.Redirect(w, r, "/web/apps?msg=missing_module", http.StatusSeeOther)
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
		http.Redirect(w, r, "/web/apps?msg=unknown_action", http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Redirect(w, r, "/web/apps?msg="+url.QueryEscape(err.Error()), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/web/apps?msg="+url.QueryEscape(msg), http.StatusSeeOther)
}
