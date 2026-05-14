package web

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"sumeru/core/mail"
	"sumeru/core/orm"
)

// ChatterPostHandler accepts POST /web/chatter/post (model, res_id, body, next).
func ChatterPostHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	if !mail.CompanyChatterEnabled(r.Context()) {
		http.Error(w, "Chatter disabled", http.StatusForbidden)
		return
	}
	modelName := strings.TrimSpace(r.PostFormValue("model"))
	next := strings.TrimSpace(r.PostFormValue("next"))
	body := strings.TrimSpace(r.PostFormValue("body"))
	if modelName == "" {
		http.Error(w, "Missing model", http.StatusBadRequest)
		return
	}
	if next == "" || !strings.HasPrefix(next, "/web") || strings.HasPrefix(next, "//") {
		next = "/web"
	}
	if body == "" {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	if utf8.RuneCountInString(body) > 10000 {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return
	}
	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		http.Error(w, "Unknown model", http.StatusBadRequest)
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
	author := "User"
	if err := mail.PostMessage(r.Context(), modelName, rid, body, mail.SubtypeComment, author); err != nil {
		log.Printf("web: chatter post %s id=%d: %v", modelName, rid, err)
		http.Error(w, "Post failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
