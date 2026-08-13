package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// SwitchCompanyPost sets the signed-in user's company_id (must exist in core.company).
func SwitchCompanyPost(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !RequirePOST(w, r) {
		return
	}
	if !ParsePostForm(w, r) {
		return
	}
	cid, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue("company_id")))
	if err != nil || cid <= 0 {
		http.Redirect(w, r, SafeWebNext(r.PostFormValue("next"), "/web/home"), http.StatusSeeOther)
		return
	}
	ctx := r.Context()
	if _, err := orm.SearchOne(ctx, "core.company", map[string]interface{}{"id": cid}); err != nil {
		http.Redirect(w, r, SafeWebNext(r.PostFormValue("next"), "/web/home"), http.StatusSeeOther)
		return
	}
	uid := SessionUserID(r)
	if uid <= 0 {
		http.Redirect(w, r, "/web/login", http.StatusSeeOther)
		return
	}
	if !orm.UserAllowedCompany(ctx, uid, int64(cid)) {
		http.Redirect(w, r, SafeWebNext(r.PostFormValue("next"), "/web/home"), http.StatusSeeOther)
		return
	}
	userTbl := orm.MustQuotedTableName("core.user")
	if _, err := orm.DB.ExecContext(ctx, `UPDATE `+userTbl+` SET company_id = $1 WHERE id = $2`, cid, uid); err != nil {
		WebLogf(ctx, "/web/company/switch", "update company_id: %v", err)
	}
	http.Redirect(w, r, SafeWebNext(r.PostFormValue("next"), "/web/home"), http.StatusSeeOther)
}
