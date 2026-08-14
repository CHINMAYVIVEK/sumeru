package web

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ImportCSVHandler imports CSV rows into a model: POST multipart model=… & file=…
func ImportCSVHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginMultipartPost(w, r, maxImportBodyBytes) {
		return
	}

	request, ok := openImportCSVRequest(w, r)
	if !ok {
		return
	}
	defer request.file.Close()

	createdCount, err := importCSVRows(r.Context(), request.modelInst, request.file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectWithWebMessage(w, r, request.next, importCSVFlashMessage(createdCount))
}

type importCSVRequest struct {
	modelInst orm.Model
	file      io.ReadCloser
	next      string
}

func openImportCSVRequest(w http.ResponseWriter, r *http.Request) (importCSVRequest, bool) {
	modelName := strings.TrimSpace(r.FormValue(importModelField))
	if modelName == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return importCSVRequest{}, false
	}

	modelInst, ok := requireRegisteredModel(w, modelName)
	if !ok {
		return importCSVRequest{}, false
	}
	if !requireModelAccess(w, r, modelName, "create") {
		return importCSVRequest{}, false
	}

	upload, _, err := r.FormFile(importFileField)
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return importCSVRequest{}, false
	}

	return importCSVRequest{
		modelInst: modelInst,
		file:      upload,
		next:      r.FormValue(nextField),
	}, true
}

func importCSVFlashMessage(createdCount int) string {
	return "imported_" + strconv.Itoa(createdCount)
}

func importCSVRows(ctx context.Context, modelInst orm.Model, file io.Reader) (int, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("empty csv")
	}
	normalizeCSVHeader(header)

	allowedFields := allowedImportFieldNames(modelInst)
	createdCount := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return createdCount, fmt.Errorf("csv error after %d rows: %v", createdCount, err)
		}

		values := importableRowValues(header, record, allowedFields)
		if len(values) == 0 {
			continue
		}
		if _, err := orm.Create(ctx, modelInst, values); err != nil {
			return createdCount, fmt.Errorf("row %d: %v", createdCount+1, err)
		}
		createdCount++
	}

	return createdCount, nil
}

func normalizeCSVHeader(header []string) {
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
}

func allowedImportFieldNames(modelInst orm.Model) map[string]struct{} {
	allowed := make(map[string]struct{})
	for _, field := range modelInst.Fields() {
		if field.Name == "" || field.Name == workspaceRecordIDParam {
			continue
		}
		allowed[field.Name] = struct{}{}
	}
	return allowed
}

func importableRowValues(header, record []string, allowedFields map[string]struct{}) map[string]interface{} {
	values := map[string]interface{}{}
	for columnIndex, columnName := range header {
		if !isImportableColumn(columnName, allowedFields) {
			continue
		}
		if columnIndex >= len(record) {
			continue
		}
		values[columnName] = coerceCSVValue(record[columnIndex])
	}
	return values
}

func isImportableColumn(columnName string, allowedFields map[string]struct{}) bool {
	columnName = strings.TrimSpace(columnName)
	if columnName == "" || columnName == workspaceRecordIDParam {
		return false
	}
	_, allowed := allowedFields[columnName]
	return allowed
}

func coerceCSVValue(raw string) interface{} {
	value := strings.TrimSpace(raw)
	if value == "" {
		return value
	}
	switch strings.ToLower(value) {
	case "true":
		return true
	case "false":
		return false
	}
	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		return number
	}
	return value
}
