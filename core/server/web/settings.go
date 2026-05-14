package web

import (
	"fmt"
	"net/http"

	"sumeru/core/orm"
)

// SettingsHomeRedirect sends /web/settings to the default Settings workspace (Companies → All Companies).
func SettingsHomeRedirect(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	mid, _, err := orm.ResolveXmlId(ctx, "base.menu_company_companies")
	if err != nil || mid == 0 {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/web?menu_id=%d", mid), http.StatusFound)
}
