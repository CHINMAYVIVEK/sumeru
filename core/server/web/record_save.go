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
)

// RecordSaveHandler applies POSTed field values to an existing row (or creates one when id is empty).
func RecordSaveHandler(w http.ResponseWriter, r *http.Request) {
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
		newID, err := orm.Create(inst, vals)
		if err != nil {
			log.Printf("web: create %s: %v", modelName, err)
			http.Error(w, "Save failed", http.StatusInternalServerError)
			return
		}
		_ = mail.PostMessage(modelName, int64(newID), fmt.Sprintf("Record created (id %d).", newID), mail.SubtypeNotification, "System")
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
	if err := orm.UpdateRecordByID(modelName, id, vals); err != nil {
		log.Printf("web: update %s id=%d: %v", modelName, id, err)
		http.Error(w, "Save failed", http.StatusInternalServerError)
		return
	}
	_ = mail.PostMessage(modelName, int64(id), fmt.Sprintf("Record updated (id %d).", id), mail.SubtypeNotification, "System")
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
		"model": {}, "id": {}, "next": {},
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
