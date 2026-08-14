package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// SwitchCompanyPost sets the signed-in user's company_id (must exist in core.company).
func SwitchCompanyPost(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}

	companyID, err := strconv.Atoi(strings.TrimSpace(r.PostFormValue(companyIDFormField)))
	if err != nil || companyID <= 0 {
		redirectToWebNext(w, r, r.PostFormValue("next"))
		return
	}

	ctx := r.Context()
	if _, err := orm.SearchOne(ctx, coreCompanyModel, map[string]interface{}{"id": companyID}); err != nil {
		redirectToWebNext(w, r, r.PostFormValue("next"))
		return
	}

	userID := SessionUserID(r)
	if userID <= 0 {
		http.Redirect(w, r, loginRoute, http.StatusSeeOther)
		return
	}
	if !orm.UserAllowedCompany(ctx, userID, int64(companyID)) {
		redirectToWebNext(w, r, r.PostFormValue("next"))
		return
	}

	userTable := orm.MustQuotedTableName(coreUserModel)
	if _, err := orm.DB.ExecContext(ctx, `UPDATE `+userTable+` SET company_id = $1 WHERE id = $2`, companyID, userID); err != nil {
		WebLogf(ctx, companySwitchRoute, "update company_id: %v", err)
	} else {
		WebLogNavigation(ctx, companySwitchRoute, "company_switch", "Active company switched", map[string]interface{}{
			"company_id": companyID,
			"user_id":    userID,
		})
	}
	redirectToWebNext(w, r, r.PostFormValue("next"))
}

func redirectToWebNext(w http.ResponseWriter, r *http.Request, rawNext string) {
	http.Redirect(w, r, SafeWebNext(rawNext, homeRoute), http.StatusSeeOther)
}
