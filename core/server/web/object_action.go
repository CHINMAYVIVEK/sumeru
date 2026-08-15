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
		redirectRecordError(w, r, homeRoute, "object_action", "", err)
		return
	}
	if !validateSessionCSRF(w, r) {
		return
	}
	if !requireLogin(w, r) {
		return
	}
	model := strings.TrimSpace(r.FormValue("model"))
	method := strings.TrimSpace(r.FormValue("method"))
	idStr := strings.TrimSpace(r.FormValue("id"))
	next := SafeWebNext(strings.TrimSpace(r.FormValue("next")), homeRoute)
	id, _ := strconv.Atoi(idStr)
	if model == "" || method == "" || id <= 0 {
		redirectRecordError(w, r, next, "object_action", model, errRequiredObjectActionFields())
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
		redirectRecordError(w, r, next, "object_action", model, err)
		return
	}
	if redirect == "" {
		redirect = next
	}
	if redirect == "" {
		redirect = "/web"
	}
	http.Redirect(w, r, SafeWebNext(redirect, "/web/home"), http.StatusSeeOther)
}

func errRequiredObjectActionFields() error {
	return errObjectActionRequired
}

var errObjectActionRequired = objectActionRequiredError("model, id, and method are required")

type objectActionRequiredError string

func (e objectActionRequiredError) Error() string { return string(e) }
