package web

import (
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
	if !requireLogin(w, r) {
		return
	}
	if !RequirePOST(w, r) {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxImportBody)
	if err := r.ParseMultipartForm(maxImportBody); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	model := strings.TrimSpace(r.FormValue("model"))
	if model == "" {
		http.Error(w, "model required", http.StatusBadRequest)
		return
	}
	if _, ok := requireRegisteredModel(w, model); !ok {
		return
	}
	ctx := r.Context()
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), model, "create"); err != nil {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil {
		http.Error(w, "empty csv", http.StatusBadRequest)
		return
	}
	for i := range header {
		header[i] = strings.TrimSpace(header[i])
	}
	inst := orm.Registry[model]
	allowed := map[string]struct{}{}
	for _, f := range inst.Fields() {
		if f.Name != "" && f.Name != "id" {
			allowed[f.Name] = struct{}{}
		}
	}
	created := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("csv error after %d rows: %v", created, err), http.StatusBadRequest)
			return
		}
		vals := map[string]interface{}{}
		for i, col := range header {
			if col == "" || col == "id" {
				continue
			}
			if _, ok := allowed[col]; !ok {
				continue
			}
			if i >= len(rec) {
				continue
			}
			vals[col] = coerceCSV(rec[i])
		}
		if len(vals) == 0 {
			continue
		}
		if _, err := orm.Create(ctx, inst, vals); err != nil {
			http.Error(w, fmt.Sprintf("row %d: %v", created+1, err), http.StatusBadRequest)
			return
		}
		created++
	}
	next := SafeWebNext(r.FormValue("next"), "/web/home")
	http.Redirect(w, r, next+"&msg=imported_"+strconv.Itoa(created), http.StatusSeeOther)
}

func coerceCSV(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	if s == "true" || s == "TRUE" {
		return true
	}
	if s == "false" || s == "FALSE" {
		return false
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	return s
}
