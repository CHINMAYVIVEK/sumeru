package web

import (
	"context"
	"fmt"
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
		redirectRecordError(w, r, SafeWebNext(r.PostFormValue(nextField), homeRoute), "record_chatter", "", fmt.Errorf("chatter disabled"))
		return
	}

	form := parseChatterPostForm(r)
	if form.Body == "" {
		redirectToWebNext(w, r, form.Next)
		return
	}

	recordID, err := validateChatterPost(r, form)
	if err != nil {
		redirectRecordError(w, r, form.Next, "record_chatter", form.ModelName, err)
		return
	}

	if err := postChatterComment(r.Context(), form.ModelName, recordID, form.Body); err != nil {
		redirectRecordError(w, r, form.Next, "record_chatter", form.ModelName, err)
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

func validateChatterPost(r *http.Request, form chatterPostForm) (recordID int64, err error) {
	if form.ModelName == "" {
		return 0, fmt.Errorf("missing model")
	}
	if chatterBodyTooLong(form.Body) {
		return 0, fmt.Errorf("message too long")
	}
	if _, ok := orm.Registry[form.ModelName]; !ok {
		return 0, fmt.Errorf("unknown model")
	}
	if form.ModelName == mailMessageModel {
		return 0, fmt.Errorf("invalid model")
	}

	recordID, err = parseChatterRecordID(form.RecordIDRaw)
	if err != nil {
		return 0, err
	}
	if !chatterTargetRecordExists(r, form.ModelName, recordID) {
		return 0, fmt.Errorf("record not found")
	}
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), form.ModelName, "write"); err != nil {
		return 0, fmt.Errorf("access denied")
	}

	return recordID, nil
}

func parseChatterRecordID(rawRecordID string) (int64, error) {
	recordID, err := strconv.ParseInt(rawRecordID, 10, 64)
	if err != nil || recordID <= 0 {
		return 0, fmt.Errorf("invalid res_id")
	}
	return recordID, nil
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
