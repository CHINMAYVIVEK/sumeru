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

	request, err := parseRecordDeleteRequest(r)
	if err != nil {
		redirectRecordError(w, r, homeRoute, "record_delete", "", err)
		return
	}

	if err := orm.Unlink(r.Context(), request.ModelName, request.RecordID); err != nil {
		redirectRecordError(w, r, workspaceListURL(request.ActionID, request.MenuID), "record_delete", request.ModelName, err)
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

func parseRecordDeleteRequest(r *http.Request) (recordDeleteRequest, error) {
	modelName := formOrQueryValue(r, recordModelField)
	recordIDRaw := formOrQueryValue(r, workspaceRecordIDParam)
	if modelName == "" || recordIDRaw == "" {
		return recordDeleteRequest{}, fmt.Errorf("missing model or id")
	}

	recordID, err := strconv.Atoi(recordIDRaw)
	if err != nil || recordID <= 0 {
		return recordDeleteRequest{}, fmt.Errorf("invalid id")
	}

	return recordDeleteRequest{
		ModelName: modelName,
		RecordID:  recordID,
		ActionID:  r.URL.Query().Get(workspaceActionParam),
		MenuID:    r.URL.Query().Get(workspaceMenuIDParam),
	}, nil
}
