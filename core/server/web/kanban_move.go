package web

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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
	if !requireLoginJSONPost(w, r) {
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<16))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	var moveRequest kanbanMoveRequest
	if err := json.Unmarshal(body, &moveRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	modelName := strings.TrimSpace(moveRequest.Model)
	fieldName := strings.TrimSpace(moveRequest.Field)
	if modelName == "" || fieldName == "" || moveRequest.RecordID <= 0 {
		http.Error(w, "Missing model, record_id, or field", http.StatusBadRequest)
		return
	}
	if _, ok := requireRegisteredModel(w, modelName); !ok {
		return
	}
	if !isMany2OneField(modelName, fieldName) {
		http.Error(w, "Field not allowed for kanban move", http.StatusBadRequest)
		return
	}

	values := map[string]interface{}{fieldName: moveRequest.Value}
	if moveRequest.Value <= 0 {
		values[fieldName] = nil
	}
	if fieldName == stageIDField {
		values[dateLastStageUpdateField] = time.Now().Format(time.RFC3339)
	}
	if err := orm.UpdateRecordByID(r.Context(), modelName, int(moveRequest.RecordID), values); err != nil {
		WebLogf(r.Context(), kanbanMoveRoute, "update %s id=%d: %v", modelName, moveRequest.RecordID, err)
		http.Error(w, "Update failed", http.StatusForbidden)
		return
	}

	writeJSONOK(w)
}

func isMany2OneField(model, fieldName string) bool {
	inst, ok := orm.Registry[model]
	if !ok {
		return false
	}
	for _, field := range inst.Fields() {
		if field.Name == fieldName && field.Type == orm.Many2One {
			return true
		}
	}
	return false
}
