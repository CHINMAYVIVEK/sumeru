package render

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

func splitFlashMessages(flashes []FlashMessage) (inline, toast []FlashMessage) {
	for _, f := range flashes {
		if f.ToastOnly {
			toast = append(toast, f)
			continue
		}
		inline = append(inline, f)
	}
	return inline, toast
}

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

func rawField(record map[string]interface{}, name string) (interface{}, bool) {
	if record == nil {
		return nil, false
	}
	v, ok := record[name]
	return v, ok
}

func fieldDef(model, fieldName string) *orm.FieldDefinition {
	inst, ok := orm.Registry[model]
	if !ok {
		return nil
	}
	for i := range inst.Fields() {
		f := inst.Fields()[i]
		if f.Name == fieldName {
			return &f
		}
	}
	return nil
}

// displayCell returns a human-readable cell value, resolving Many2One to display name.
func displayCell(ctx context.Context, model, fieldName string, row map[string]interface{}) string {
	if fd := fieldDef(model, fieldName); fd != nil && fd.Type == orm.Many2One {
		if id, ok := orm.CoerceInt64(row[fieldName]); ok && id > 0 {
			if n := orm.DisplayNameForID(ctx, fd.Relation, int(id)); n != "" {
				return n
			}
		}
		return ""
	}
	return strings.TrimSpace(recStr(row, fieldName))
}
