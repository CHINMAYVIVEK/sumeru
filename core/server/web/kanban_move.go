package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

type kanbanMoveRequest struct {
	Model     string `json:"model"`
	RecordID  int64  `json:"record_id"`
	Field     string `json:"field"`
	Value     int64  `json:"value"`
	CSRFToken string `json:"csrf_token"`
}

// KanbanMoveHandler updates a grouped kanban field (e.g. stage_id) after drag-and-drop.
func KanbanMoveHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !validateSessionCSRF(w, r) {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var req kanbanMoveRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	modelName := strings.TrimSpace(req.Model)
	fieldName := strings.TrimSpace(req.Field)
	if modelName == "" || fieldName == "" || req.RecordID <= 0 {
		http.Error(w, "Missing model, record_id, or field", http.StatusBadRequest)
		return
	}
	if _, ok := orm.Registry[modelName]; !ok {
		http.Error(w, "Unknown model", http.StatusBadRequest)
		return
	}
	if !isMany2OneField(modelName, fieldName) {
		http.Error(w, "Field not allowed for kanban move", http.StatusBadRequest)
		return
	}

	vals := map[string]interface{}{fieldName: req.Value}
	if req.Value <= 0 {
		vals[fieldName] = nil
	}
	if fieldName == "stage_id" {
		vals["date_last_stage_update"] = time.Now().Format(time.RFC3339)
	}
	if err := orm.UpdateRecordByID(r.Context(), modelName, int(req.RecordID), vals); err != nil {
		WebLogf(r.Context(), "/web/kanban/move", "update %s id=%d: %v", modelName, req.RecordID, err)
		http.Error(w, "Update failed", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func isMany2OneField(model, fieldName string) bool {
	inst, ok := orm.Registry[model]
	if !ok {
		return false
	}
	for _, f := range inst.Fields() {
		if f.Name == fieldName && f.Type == orm.Many2One {
			return true
		}
	}
	return false
}

// KanbanMoveFieldAllowed checks whether a field is the active kanban group field for a view arch.
func KanbanMoveFieldAllowed(arch, fieldName string) bool {
	v, err := parser.ParseViewFromArch(arch)
	if err != nil {
		return false
	}
	return v.KanbanGroupField() == strings.TrimSpace(fieldName)
}

// ParseKanbanMoveValue parses group value from form/JSON (allows empty → 0).
func ParseKanbanMoveValue(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" || s == "false" || s == "null" {
		return 0
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
