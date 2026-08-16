package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/addons/mail"
	"sumeru/core/orm"
)

// RecordSaveHandler applies POSTed field values to an existing row (or creates one when id is empty).
func RecordSaveHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginAndPOST(w, r) {
		return
	}

	form := parseRecordSaveForm(r)
	if form.ModelName == "" {
		redirectRecordError(w, r, form.Next, "record_save", "", fmt.Errorf("missing model"))
		return
	}

	modelInst, ok := requireRegisteredModel(w, form.ModelName)
	if !ok {
		return
	}

	fieldValues, err := postFormToModelValues(modelInst, r.PostForm)
	if err != nil {
		next := ensureFormEditRedirectURL(form.Next, isNewRecord(form.RecordIDRaw))
		redirectRecordError(w, r, next, "record_save", form.ModelName, err)
		return
	}

	if isNewRecord(form.RecordIDRaw) {
		handleRecordCreate(w, r, form, modelInst, fieldValues)
		return
	}
	handleRecordUpdate(w, r, form, modelInst, fieldValues)
}

type recordSaveForm struct {
	ModelName   string
	RecordIDRaw string
	Next        string
}

func parseRecordSaveForm(r *http.Request) recordSaveForm {
	return recordSaveForm{
		ModelName:   strings.TrimSpace(r.PostFormValue(recordModelField)),
		RecordIDRaw: strings.TrimSpace(r.PostFormValue(workspaceRecordIDParam)),
		Next:        SafeWebNext(r.PostFormValue(nextField), homeRoute),
	}
}

func isNewRecord(recordIDRaw string) bool {
	return recordIDRaw == "" || recordIDRaw == "0"
}

func handleRecordCreate(w http.ResponseWriter, r *http.Request, form recordSaveForm, modelInst orm.Model, fieldValues map[string]interface{}) {
	ctx := r.Context()
	applyActionDefaultsOnCreate(ctx, form.Next, fieldValues)
	applyCreateOwnershipDefaults(ctx, modelInst, fieldValues)
	newRecordID, err := orm.Create(ctx, modelInst, fieldValues)
	if err != nil {
		next := ensureFormEditRedirectURL(form.Next, true)
		redirectRecordError(w, r, next, "record_save", form.ModelName, err)
		return
	}

	applyUserSecurityPostIfNeeded(ctx, form.ModelName, newRecordID, r.PostForm)
	notifyRecordSaved(ctx, form.ModelName, newRecordID, fmt.Sprintf("Record created (id %d).", newRecordID))
	redirectRecordSuccess(w, r, workspaceFormURL(form.Next, newRecordID), saveOKCreatedMsg)
}

func handleRecordUpdate(w http.ResponseWriter, r *http.Request, form recordSaveForm, modelInst orm.Model, fieldValues map[string]interface{}) {
	recordID, err := strconv.Atoi(form.RecordIDRaw)
	if err != nil || recordID <= 0 {
		next := ensureFormEditRedirectURL(form.Next, false)
		redirectRecordError(w, r, next, "record_save", form.ModelName, fmt.Errorf("invalid id"))
		return
	}

	ctx := r.Context()
	if err := orm.UpdateRecordByID(ctx, form.ModelName, recordID, fieldValues); err != nil {
		next := ensureFormEditRedirectURL(form.Next, false)
		redirectRecordError(w, r, next, "record_save", form.ModelName, err)
		return
	}

	applyUserSecurityPostIfNeeded(ctx, form.ModelName, recordID, r.PostForm)
	notifyRecordSaved(ctx, form.ModelName, recordID, fmt.Sprintf("Record updated (id %d).", recordID))
	redirectRecordSuccess(w, r, form.Next, saveOKUpdatedMsg)
}

func applyUserSecurityPostIfNeeded(ctx context.Context, modelName string, recordID int, form url.Values) {
	if modelName != coreUserModel {
		return
	}
	orm.ApplyUserSecurityPost(ctx, orm.SecurityUID(ctx), recordID, form)
}

func notifyRecordSaved(ctx context.Context, modelName string, recordID int, message string) {
	_ = mail.PostMessage(ctx, modelName, int64(recordID), message, mail.SubtypeNotification, recordSaveSystemAuthor)
}

func applyActionDefaultsOnCreate(ctx context.Context, nextURL string, fieldValues map[string]interface{}) {
	if len(fieldValues) == 0 {
		return
	}
	actionID := actionIDFromNextURL(nextURL)
	if actionID <= 0 {
		return
	}
	actionData, err := loadWindowAction(ctx, actionID)
	if err != nil {
		return
	}
	for key, val := range actionDefaultFieldValues(actionData) {
		if _, ok := fieldValues[key]; !ok {
			fieldValues[key] = val
		}
	}
}

func actionIDFromNextURL(nextURL string) int {
	parsed, err := url.Parse(nextURL)
	if err != nil {
		return 0
	}
	raw := strings.TrimSpace(parsed.Query().Get(workspaceActionParam))
	if raw == "" {
		return 0
	}
	actionID, err := strconv.Atoi(raw)
	if err != nil || actionID <= 0 {
		return 0
	}
	return actionID
}

func workspaceFormURL(next string, recordID int) string {
	parsed, err := url.Parse(next)
	if err != nil {
		query := url.Values{}
		setWorkspaceQueryString(query, workspaceViewTypeParam, workspaceViewModeForm)
		setWorkspaceQueryInt(query, workspaceRecordIDParam, recordID)
		return workspaceRoute + "?" + query.Encode()
	}

	query := parsed.Query()
	query.Set(workspaceViewTypeParam, workspaceViewModeForm)
	query.Set(workspaceRecordIDParam, strconv.Itoa(recordID))
	query.Del(workspaceEditParam)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func postFormToModelValues(modelInst orm.Model, form url.Values) (map[string]interface{}, error) {
	fieldTypes := fieldTypesByName(modelInst)
	values := make(map[string]interface{})

	for fieldName, rawValues := range form {
		if isSkippedSaveFormField(fieldName) {
			continue
		}
		fieldType, known := fieldTypes[fieldName]
		if !known || len(rawValues) == 0 {
			continue
		}

		coerced, skip, err := coerceSaveFieldValue(fieldName, fieldType, strings.TrimSpace(rawValues[0]))
		if err != nil {
			return nil, err
		}
		if skip {
			continue
		}
		values[fieldName] = coerced
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no valid fields to save")
	}
	return values, nil
}

func fieldTypesByName(modelInst orm.Model) map[string]orm.FieldType {
	types := make(map[string]orm.FieldType, len(modelInst.Fields()))
	for _, field := range modelInst.Fields() {
		types[field.Name] = field.Type
	}
	return types
}

func isSkippedSaveFormField(fieldName string) bool {
	switch fieldName {
	case recordModelField, workspaceRecordIDParam, nextField,
		passwordPlainField, securityGroupIDsField, securityGroupsTouchedField,
		securityUserTypeField, companyIDsField:
		return true
	default:
		return false
	}
}

func coerceSaveFieldValue(fieldName string, fieldType orm.FieldType, rawValue string) (value interface{}, skip bool, err error) {
	switch fieldType {
	case orm.Boolean:
		return rawValue == "on" || rawValue == "1" || strings.EqualFold(rawValue, "true"), false, nil
	case orm.Integer, orm.Many2One:
		if rawValue == "" {
			return nil, true, nil
		}
		number, parseErr := strconv.ParseInt(rawValue, 10, 64)
		if parseErr != nil {
			return nil, false, fmt.Errorf("invalid %s", fieldName)
		}
		return int(number), false, nil
	case orm.Float, orm.Numeric:
		if rawValue == "" {
			return nil, true, nil
		}
		number, parseErr := strconv.ParseFloat(rawValue, 64)
		if parseErr != nil {
			return nil, false, fmt.Errorf("invalid %s", fieldName)
		}
		return number, false, nil
	case orm.Many2Many:
		return nil, true, nil
	case orm.Date, orm.DateTime:
		if rawValue == "" {
			return nil, false, nil
		}
		return rawValue, false, nil
	default:
		return rawValue, false, nil
	}
}
