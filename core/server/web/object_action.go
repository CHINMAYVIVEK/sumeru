package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ActionObjectHandler POST /web/action/object — runs a registered object button handler.
// Form fields: model, id, method, next (optional redirect fallback).
func ActionObjectHandler(w http.ResponseWriter, r *http.Request) {
	if !RequirePOST(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	model := strings.TrimSpace(r.FormValue("model"))
	method := strings.TrimSpace(r.FormValue("method"))
	idStr := strings.TrimSpace(r.FormValue("id"))
	next := strings.TrimSpace(r.FormValue("next"))
	id, _ := strconv.Atoi(idStr)
	if model == "" || method == "" || id <= 0 {
		http.Error(w, "model, id, and method are required", http.StatusBadRequest)
		return
	}
	vals := map[string]string{}
	for k, vs := range r.Form {
		if len(vs) > 0 {
			vals[k] = vs[0]
		}
	}
	ctx := r.Context()
	redirect, err := orm.RunObjectAction(ctx, model, id, method, vals)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if redirect == "" {
		redirect = next
	}
	if redirect == "" {
		redirect = "/web"
	}
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}
