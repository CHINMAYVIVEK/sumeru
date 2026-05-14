package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/mail"
	"sumeru/core/orm"

	"golang.org/x/crypto/bcrypt"
)

// RecordSaveHandler applies POSTed field values to an existing row (or creates one when id is empty).
func RecordSaveHandler(w http.ResponseWriter, r *http.Request) {
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
	modelName := strings.TrimSpace(r.PostFormValue("model"))
	next := strings.TrimSpace(r.PostFormValue("next"))
	if modelName == "" {
		http.Error(w, "Missing model", http.StatusBadRequest)
		return
	}
	if next == "" || !strings.HasPrefix(next, "/web") || strings.HasPrefix(next, "//") {
		next = "/web"
	}

	inst, ok := orm.Registry[modelName]
	if !ok || inst == nil {
		http.Error(w, "Unknown model", http.StatusBadRequest)
		return
	}

	idStr := strings.TrimSpace(r.PostFormValue("id"))
	if idStr == "" || idStr == "0" {
		vals, err := postFormToModelValues(inst, r.PostForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		newID, err := orm.Create(r.Context(), inst, vals)
		if err != nil {
			log.Printf("web: create %s: %v", modelName, err)
			http.Error(w, "Save failed", http.StatusInternalServerError)
			return
		}
		applyResUsersSecurityPost(r, modelName, newID)
		_ = mail.PostMessage(r.Context(), modelName, int64(newID), fmt.Sprintf("Record created (id %d).", newID), mail.SubtypeNotification, "System")
		redir := appendRecordIDToNext(next, newID)
		http.Redirect(w, r, redir, http.StatusSeeOther)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	vals, err := postFormToModelValues(inst, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := orm.UpdateRecordByID(r.Context(), modelName, id, vals); err != nil {
		log.Printf("web: update %s id=%d: %v", modelName, id, err)
		http.Error(w, "Save failed", http.StatusInternalServerError)
		return
	}
	applyResUsersSecurityPost(r, modelName, id)
	_ = mail.PostMessage(r.Context(), modelName, int64(id), fmt.Sprintf("Record updated (id %d).", id), mail.SubtypeNotification, "System")
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func appendRecordIDToNext(next string, newID int) string {
	u, err := url.Parse(next)
	if err != nil {
		return fmt.Sprintf("/web?view_type=form&id=%d", newID)
	}
	q := u.Query()
	q.Set("view_type", "form")
	q.Set("id", fmt.Sprintf("%d", newID))
	q.Del("edit")
	u.RawQuery = q.Encode()
	return u.String()
}

func postFormToModelValues(inst orm.Model, form url.Values) (map[string]interface{}, error) {
	typBy := map[string]orm.FieldType{}
	for _, f := range inst.Fields() {
		typBy[f.Name] = f.Type
	}
	skipNames := map[string]struct{}{
		"model": {}, "id": {}, "next": {}, "password_plain": {}, "security_group_ids": {}, "security_groups_touched": {},
	}
	out := make(map[string]interface{})
	for k, vv := range form {
		if _, skip := skipNames[k]; skip {
			continue
		}
		ft, ok := typBy[k]
		if !ok {
			continue
		}
		if len(vv) == 0 {
			continue
		}
		s := strings.TrimSpace(vv[0])
		switch ft {
		case orm.Boolean:
			out[k] = s == "on" || s == "1" || strings.EqualFold(s, "true")
		case orm.Integer, orm.Many2One:
			if s == "" {
				continue
			}
			n, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid %s", k)
			}
			out[k] = int(n)
		case orm.Float, orm.Numeric:
			if s == "" {
				continue
			}
			x, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid %s", k)
			}
			out[k] = x
		default:
			out[k] = s
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no valid fields to save")
	}
	return out, nil
}

func applyResUsersSecurityPost(r *http.Request, modelName string, userID int) {
	if modelName != "res.users" || userID <= 0 {
		return
	}
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), "res.users", "write"); err != nil {
		return
	}
	if r.PostFormValue("security_groups_touched") == "1" {
		var gids []int
		for _, s := range r.Form["security_group_ids"] {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			if err == nil && n > 0 {
				gids = append(gids, n)
			}
		}
		if err := orm.SetUserGroupLinks(r.Context(), userID, gids); err != nil {
			log.Printf("web: set user %d groups: %v", userID, err)
		}
	}
	if _, ok := r.Form["password_plain"]; ok {
		if pw := strings.TrimSpace(r.PostFormValue("password_plain")); pw != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("web: bcrypt: %v", err)
				return
			}
			tbl := orm.GetTableName("res.users")
			if _, err := orm.DB.ExecContext(r.Context(), `UPDATE `+tbl+` SET password = $1 WHERE id = $2`, string(hash), userID); err != nil {
				log.Printf("web: password update user %d: %v", userID, err)
			}
		}
	}
}
