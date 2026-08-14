package web

import (
	"context"
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

	form := parseChatterPostForm(r)
	if form.Body == "" {
		redirectToWebNext(w, r, form.Next)
		return
	}

	recordID, ok := validateChatterPost(w, r, form)
	if !ok {
		return
	}

	if err := postChatterComment(r.Context(), form.ModelName, recordID, form.Body); err != nil {
		WebLogf(r.Context(), chatterPostRoute, "chatter post %s id=%d: %v", form.ModelName, recordID, err)
		http.Error(w, "Post failed", http.StatusInternalServerError)
		return
	}

	redirectToWebNext(w, r, form.Next)
}

type chatterPostForm struct {
	ModelName   string
	RecordIDRaw string
	Body        string
	Next        string
}

func parseChatterPostForm(r *http.Request) chatterPostForm {
	return chatterPostForm{
		ModelName:   strings.TrimSpace(r.PostFormValue(recordModelField)),
		RecordIDRaw: strings.TrimSpace(r.PostFormValue(chatterRecordIDField)),
		Body:        strings.TrimSpace(r.PostFormValue(chatterBodyField)),
		Next:        SafeWebNext(r.PostFormValue(nextField), homeRoute),
	}
}

func validateChatterPost(w http.ResponseWriter, r *http.Request, form chatterPostForm) (recordID int64, ok bool) {
	if form.ModelName == "" {
		http.Error(w, "Missing model", http.StatusBadRequest)
		return 0, false
	}
	if chatterBodyTooLong(form.Body) {
		http.Error(w, "Message too long", http.StatusBadRequest)
		return 0, false
	}
	if _, registered := requireRegisteredModel(w, form.ModelName); !registered {
		return 0, false
	}
	if form.ModelName == mailMessageModel {
		http.Error(w, "Invalid model", http.StatusBadRequest)
		return 0, false
	}

	recordID, parseOK := parseChatterRecordID(w, form.RecordIDRaw)
	if !parseOK {
		return 0, false
	}
	if !chatterTargetRecordExists(r, form.ModelName, recordID) {
		http.Error(w, "Record not found", http.StatusNotFound)
		return 0, false
	}
	if !requireModelAccess(w, r, form.ModelName, "write") {
		return 0, false
	}

	return recordID, true
}

func parseChatterRecordID(w http.ResponseWriter, rawRecordID string) (int64, bool) {
	recordID, err := strconv.ParseInt(rawRecordID, 10, 64)
	if err != nil || recordID <= 0 {
		http.Error(w, "Invalid res_id", http.StatusBadRequest)
		return 0, false
	}
	return recordID, true
}

func chatterTargetRecordExists(r *http.Request, modelName string, recordID int64) bool {
	_, err := orm.SearchOne(r.Context(), modelName, map[string]interface{}{"id": int(recordID)})
	return err == nil
}

func chatterBodyTooLong(body string) bool {
	return utf8.RuneCountInString(body) > maxChatterBodyRunes
}

func postChatterComment(ctx context.Context, modelName string, recordID int64, body string) error {
	return mail.PostMessage(ctx, modelName, recordID, body, mail.SubtypeComment, chatterDefaultAuthor)
}
