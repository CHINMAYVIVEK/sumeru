package web

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"sumeru/addons/mail"
	"sumeru/core/orm"
)

// ChatterPostHandler accepts POST /web/chatter/post (model, res_id, body, next).
func ChatterPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}
	if !mail.CompanyChatterEnabled(r.Context()) {
		http.Error(w, "Chatter disabled", http.StatusForbidden)
		return
	}
	modelName := strings.TrimSpace(r.PostFormValue("model"))
	next := SafeWebNext(r.PostFormValue("next"), "/web/home")
	body := strings.TrimSpace(r.PostFormValue("body"))
	if modelName == "" {
		http.Error(w, "Missing model", http.StatusBadRequest)
		return
	}
	if body == "" {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	if utf8.RuneCountInString(body) > 10000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}
	if _, ok := requireRegisteredModel(w, modelName); !ok {
		return
	}
	if modelName == "mail.message" {
		http.Error(w, "Invalid model", http.StatusBadRequest)
		return
	}
	idStr := strings.TrimSpace(r.PostFormValue("res_id"))
	rid, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || rid <= 0 {
		http.Error(w, "Invalid res_id", http.StatusBadRequest)
		return
	}
	if _, err := orm.SearchOne(r.Context(), modelName, map[string]interface{}{"id": int(rid)}); err != nil {
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), modelName, "write"); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	author := "User"
	if err := mail.PostMessage(r.Context(), modelName, rid, body, mail.SubtypeComment, author); err != nil {
		WebLogf(r.Context(), "/web/chatter/post", "chatter post %s id=%d: %v", modelName, rid, err)
		http.Error(w, "Post failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
