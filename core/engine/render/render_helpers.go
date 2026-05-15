package render

import (
	"fmt"
	"net/url"
	"strings"

	"sumeru/core/orm"
)

func recStr(rec map[string]interface{}, name string) string {
	if rec == nil {
		return ""
	}
	return strings.TrimSpace(orm.AsString(rec[name]))
}

func isTruthyDB(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case int32:
		return t != 0
	case int:
		return t != 0
	case float64:
		return t != 0
	case []byte:
		s := strings.ToLower(strings.TrimSpace(string(t)))
		return s == "t" || s == "true" || s == "1"
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "t" || s == "true" || s == "1"
	default:
		return false
	}
}

func rowOpenURL(actionID int, menuID string, rowID int64) string {
	q := url.Values{}
	if actionID > 0 {
		q.Set("action", fmt.Sprintf("%d", actionID))
	}
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	q.Set("view_type", "form")
	q.Set("id", fmt.Sprintf("%d", rowID))
	return "/web?" + q.Encode()
}

func formFieldReadonly(vr *ViewRecordData) bool {
	if vr == nil || strings.TrimSpace(vr.ResModel) == "" {
		return true
	}
	if vr.RecordID == 0 {
		return false
	}
	return !vr.FormEditing
}

func workspaceFormChrome(vr *ViewRecordData) bool {
	return vr != nil && strings.TrimSpace(vr.ResModel) != ""
}

func formNewRecordURL(actionID int, menuID string) string {
	q := url.Values{}
	if actionID > 0 {
		q.Set("action", fmt.Sprintf("%d", actionID))
	}
	if strings.TrimSpace(menuID) != "" {
		q.Set("menu_id", strings.TrimSpace(menuID))
	}
	q.Set("view_type", "form")
	return "/web?" + q.Encode()
}

func rawField(record map[string]interface{}, name string) (interface{}, bool) {
	if record == nil {
		return nil, false
	}
	v, ok := record[name]
	return v, ok
}
