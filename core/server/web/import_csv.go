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

const maxImportBody = 8 << 20

// ImportCSVHandler imports CSV rows into a model: POST multipart model=… & file=…
func ImportCSVHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLoginMultipartPost(w, r, maxImportBody) {
		return
	}

	modelName := strings.TrimSpace(r.FormValue("model"))
	if modelName == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	modelInst, ok := requireRegisteredModel(w, modelName)
	if !ok {
		return
	}
	ctx := r.Context()
	if !requireModelAccess(w, r, modelName, "create") {
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	created, err := importCSVRows(ctx, modelInst, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	redirectWithWebMessage(w, r, r.FormValue("next"), "imported_"+strconv.Itoa(created))
}

func importCSVRows(ctx context.Context, modelInst orm.Model, file io.Reader) (int, error) {
	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("empty csv")
	}
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}

	allowedFields := buildAllowedImportFields(modelInst)
	created := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return created, fmt.Errorf("csv error after %d rows: %v", created, err)
		}
		values := map[string]interface{}{}
		for i, column := range header {
			if column == "" || column == "id" {
				continue
			}
			if _, ok := allowedFields[column]; !ok {
				continue
			}
			if i >= len(record) {
				continue
			}
			values[column] = coerceCSV(record[i])
		}
		if len(values) == 0 {
			continue
		}
		if _, err := orm.Create(ctx, modelInst, values); err != nil {
			return created, fmt.Errorf("row %d: %v", created+1, err)
		}
		created++
	}
	return created, nil
}

func buildAllowedImportFields(modelInst orm.Model) map[string]struct{} {
	allowed := map[string]struct{}{}
	for _, field := range modelInst.Fields() {
		if field.Name != "" && field.Name != "id" {
			allowed[field.Name] = struct{}{}
		}
	}
	return allowed
}

func coerceCSV(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if value == "true" || value == "TRUE" {
		return true
	}
	if value == "false" || value == "FALSE" {
		return false
	}
	if number, err := strconv.ParseInt(value, 10, 64); err == nil {
		return number
	}
	return value
}
