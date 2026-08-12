package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ActionCreateAPIKey generates a one-time raw API key for a user (POST user_id, name).
// The raw key is returned once via the redirect query param api_key (show immediately in UI).
func ActionCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	uid, _ := strconv.Atoi(strings.TrimSpace(r.PostFormValue("user_id")))
	if uid <= 0 {
		uid = SessionUserID(r)
	}
	name := strings.TrimSpace(r.PostFormValue("name"))
	ctx := r.Context()
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "core.user.apikey", "create"); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	raw, err := orm.CreateAPIKeyForUser(ctx, uid, name)
	if err != nil {
		WebLogf(ctx, "/web/action/create_api_key", "%v", err)
		http.Error(w, "Could not create API key", http.StatusInternalServerError)
		return
	}
	next := SafeWebNext(r.PostFormValue("next"), "/web/home")
	sep := "?"
	if strings.Contains(next, "?") {
		sep = "&"
	}
	http.Redirect(w, r, next+sep+"msg=api_key_created&api_key="+url.QueryEscape(raw), http.StatusSeeOther)
}
