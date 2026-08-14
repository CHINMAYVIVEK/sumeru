package web

import (
	"fmt"
	"net/http"
	"strconv"

	"sumeru/core/orm"
)

// RecordDeleteHandler removes one row and redirects back to the workspace list view.
func RecordDeleteHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}

	request, ok := parseRecordDeleteRequest(w, r)
	if !ok {
		return
	}

	if err := orm.Unlink(r.Context(), request.ModelName, request.RecordID); err != nil {
		http.Error(w, fmt.Sprintf("Delete failed: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, workspaceListURL(request.ActionID, request.MenuID), http.StatusSeeOther)
}

type recordDeleteRequest struct {
	ModelName string
	RecordID  int
	ActionID  string
	MenuID    string
}

func parseRecordDeleteRequest(w http.ResponseWriter, r *http.Request) (recordDeleteRequest, bool) {
	modelName := formOrQueryValue(r, recordModelField)
	recordIDRaw := formOrQueryValue(r, workspaceRecordIDParam)
	if modelName == "" || recordIDRaw == "" {
		http.Error(w, "Missing model or id", http.StatusBadRequest)
		return recordDeleteRequest{}, false
	}

	recordID, err := strconv.Atoi(recordIDRaw)
	if err != nil || recordID <= 0 {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return recordDeleteRequest{}, false
	}

	return recordDeleteRequest{
		ModelName: modelName,
		RecordID:  recordID,
		ActionID:  r.URL.Query().Get(workspaceActionParam),
		MenuID:    r.URL.Query().Get(workspaceMenuIDParam),
	}, true
}
