package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

func RecordDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !RequirePOST(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	if !validateSessionCSRF(w, r) {
		return
	}

	modelName := strings.TrimSpace(r.FormValue("model"))
	if modelName == "" {
		modelName = strings.TrimSpace(r.URL.Query().Get("model"))
	}
	idStr := strings.TrimSpace(r.FormValue("id"))
	if idStr == "" {
		idStr = strings.TrimSpace(r.URL.Query().Get("id"))
	}
	if modelName == "" || idStr == "" {
		http.Error(w, "Missing model or id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	if err := orm.Unlink(r.Context(), modelName, id); err != nil {
		http.Error(w, fmt.Sprintf("Delete failed: %v", err), http.StatusInternalServerError)
		return
	}

	actionID := r.URL.Query().Get("action")
	menuID := r.URL.Query().Get("menu_id")
	redirectURL := fmt.Sprintf("/web?action=%s&menu_id=%s", actionID, menuID)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
