package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/module"
	"sumeru/core/orm"
)

// ModuleActionHandler handles POST actions: install, uninstall, activate, deactivate, save.
func ModuleActionHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !requireModelAccess(w, r, appsModuleModel, "write") {
		return
	}

	browse := parseAppsBrowseStateFromForm(r)
	action := strings.TrimSpace(r.FormValue("do"))
	moduleName := strings.TrimSpace(r.FormValue("module"))
	if moduleName == "" {
		http.Redirect(w, r, appsRedirectURL("missing_module", browse), http.StatusSeeOther)
		return
	}

	switch action {
	case moduleActionSaveModule:
		if err := saveModuleFromForm(r); err != nil {
			http.Redirect(w, r, appsRedirectURL(err.Error(), browse), http.StatusSeeOther)
			return
		}
		saveBrowse := browse
		saveBrowse.ModuleName = moduleName
		http.Redirect(w, r, appsDetailRedirectURL("saved", saveBrowse), http.StatusSeeOther)
		return
	case moduleActionInstall, moduleActionUninstall, moduleActionDeactivate, moduleActionActivate:
		message, err := runModuleLifecycleAction(r.Context(), action, moduleName)
		if err != nil {
			http.Redirect(w, r, appsRedirectURL(err.Error(), browse), http.StatusSeeOther)
			return
		}
		WebLogNavigation(r.Context(), moduleActionRoute, "module_action", "Module action completed", map[string]interface{}{
			"do":     action,
			"module": moduleName,
		})
		http.Redirect(w, r, appsRedirectURL(message, browse), http.StatusSeeOther)
		return
	default:
		http.Redirect(w, r, appsRedirectURL("unknown_action", browse), http.StatusSeeOther)
	}
}

// appsDetailRedirectURL redirects back to a module detail view after save.
func appsDetailRedirectURL(message string, browse appsBrowseState) string {
	query := url.Values{}
	if strings.TrimSpace(message) != "" {
		query.Set("msg", strings.TrimSpace(message))
	}
	appendAppsQueryBase(query, browse.Layout, browse.Filter, browse.Scope, browse.SearchQuery)
	query.Set("module", browse.ModuleName)
	return appsRoute + "?" + query.Encode()
}

func runModuleLifecycleAction(ctx context.Context, action, moduleName string) (message string, err error) {
	switch action {
	case moduleActionInstall:
		err = module.InstallModuleByName(ctx, moduleName)
		message = "installed_" + moduleName
	case moduleActionUninstall:
		err = module.UninstallModuleByName(ctx, moduleName)
		message = "uninstalled_" + moduleName
	case moduleActionDeactivate:
		err = module.SetModuleActive(ctx, moduleName, false)
		message = "deactivated_" + moduleName
	case moduleActionActivate:
		err = module.SetModuleActive(ctx, moduleName, true)
		message = "activated_" + moduleName
	default:
		return "", fmt.Errorf("unknown_action")
	}
	return message, err
}

func saveModuleFromForm(r *http.Request) error {
	moduleName := strings.TrimSpace(r.FormValue("module"))
	if moduleName == "" {
		return fmt.Errorf("missing_module")
	}
	recordID, err := strconv.Atoi(strings.TrimSpace(r.FormValue("module_row_id")))
	if err != nil || recordID <= 0 {
		return fmt.Errorf("invalid_module_row")
	}
	row, err := orm.SearchOne(r.Context(), appsModuleModel, map[string]interface{}{"id": recordID})
	if err != nil {
		return fmt.Errorf("module_not_found")
	}
	if strings.TrimSpace(orm.AsString(row["name"])) != moduleName {
		return fmt.Errorf("module_mismatch")
	}
	values := map[string]interface{}{
		"display_name": strings.TrimSpace(r.FormValue("display_name")),
		"author":       strings.TrimSpace(r.FormValue("author")),
		"description":  strings.TrimSpace(r.FormValue("description")),
	}
	return orm.UpdateRecordByID(r.Context(), appsModuleModel, recordID, values)
}
