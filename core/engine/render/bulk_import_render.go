package render

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/report"
)

func init() {
	RegisterNotebookHook("sys.bulk.import", "column mapping", renderBulkImportMapping)
	RegisterNotebookHook("sys.bulk.import", "preview", renderBulkImportPreview)
}

func renderBulkImportMapping(ctx context.Context, vr *ViewRecordData, _ bool) template.HTML {
	if vr == nil || vr.RecordID <= 0 {
		return template.HTML(`<p class="sum-bulk-import-empty">Save the batch first.</p>`)
	}
	record := vr.Record
	if record == nil {
		record = map[string]interface{}{}
	}
	targetModel := strings.TrimSpace(orm.AsString(record["target_model"]))
	headers := parseJSONStringArray(orm.AsString(record["csv_headers"]))
	mapping := parseJSONStringMap(orm.AsString(record["column_mapping"]))
	fieldOptions := bulkImportFieldOptions(targetModel)

	var sb strings.Builder
	sb.WriteString(`<div class="sum-bulk-import-form">`)
	sb.WriteString(`<input type="hidden" name="column_mapping" value="` + template.HTMLEscapeString(orm.AsString(record["column_mapping"])) + `" />`)
	sb.WriteString(`<table class="sum-bulk-map-table"><thead><tr><th>CSV column</th><th>Model field</th></tr></thead><tbody>`)
	for _, h := range headers {
		sb.WriteString(`<tr class="sum-bulk-map-row" data-csv-header="` + template.HTMLEscapeString(h) + `">`)
		sb.WriteString(`<td>` + template.HTMLEscapeString(h) + `</td><td>`)
		sb.WriteString(`<select class="sum-bulk-map-select">`)
		sb.WriteString(`<option value="">Skip column</option>`)
		sb.WriteString(`<option value="-">— Skip —</option>`)
		for _, opt := range fieldOptions {
			selected := ""
			if mapping[h] == opt.Name {
				selected = ` selected`
			}
			sb.WriteString(`<option value="` + template.HTMLEscapeString(opt.Name) + `"` + selected + `>`)
			sb.WriteString(template.HTMLEscapeString(opt.Label) + `</option>`)
		}
		sb.WriteString(`</select></td></tr>`)
	}
	sb.WriteString(`</tbody></table></div>`)
	return template.HTML(sb.String())
}

func renderBulkImportPreview(ctx context.Context, vr *ViewRecordData, _ bool) template.HTML {
	if vr == nil || vr.RecordID <= 0 {
		return template.HTML(`<p class="sum-bulk-import-empty">No preview available.</p>`)
	}
	mapping := parseJSONStringMap(orm.AsString(vr.Record["column_mapping"]))
	preview, err := report.PreviewBulkImport(ctx, report.PreviewBulkImportInput{
		BatchID: vr.RecordID,
		Mapping: mapping,
	})
	if err != nil {
		return template.HTML(`<p class="sum-bulk-import-error">` + template.HTMLEscapeString(err.Error()) + `</p>`)
	}

	var sb strings.Builder
	sb.WriteString(`<div class="sum-bulk-preview">`)
	sb.WriteString(fmt.Sprintf(`<p class="sum-bulk-preview-summary">Showing %d of %d rows — %d validation issue(s)</p>`,
		len(preview.Rows), preview.TotalRows, preview.ErrorCount))
	sb.WriteString(`<div class="sum-web-table-wrap"><table class="sum-list-table sum-bulk-preview-table"><thead><tr>`)
	if len(preview.Rows) > 0 {
		cols := sortedKeys(preview.Rows[0].Values)
		for _, field := range cols {
			sb.WriteString(`<th>` + template.HTMLEscapeString(field) + `</th>`)
		}
		sb.WriteString(`<th>Errors</th></tr></thead><tbody>`)
		for _, row := range preview.Rows {
			sb.WriteString(`<tr>`)
			for _, field := range cols {
				val := template.HTMLEscapeString(orm.AsString(row.Values[field]))
				sb.WriteString(`<td>` + val + `</td>`)
			}
			errText := strings.Join(row.Errors, "; ")
			cellClass := ""
			if errText != "" {
				cellClass = ` class="sum-bulk-preview-error"`
			}
			sb.WriteString(`<td` + cellClass + `>` + template.HTMLEscapeString(errText) + `</td></tr>`)
		}
	} else {
		sb.WriteString(`<th>Row</th><th>Status</th></tr></thead><tbody>`)
		sb.WriteString(`<tr><td colspan="2">No data rows in CSV</td></tr>`)
	}
	sb.WriteString(`</tbody></table></div></div>`)
	return template.HTML(sb.String())
}

type bulkFieldOption struct {
	Name  string
	Label string
}

func bulkImportFieldOptions(modelName string) []bulkFieldOption {
	modelName = strings.TrimSpace(modelName)
	modelInst, ok := orm.Registry[modelName]
	if !ok {
		return nil
	}
	var out []bulkFieldOption
	for _, f := range modelInst.Fields() {
		if f.Name == "" {
			continue
		}
		label := f.String
		if label == "" {
			label = f.Name
		}
		out = append(out, bulkFieldOption{Name: f.Name, Label: label})
	}
	return out
}

func parseJSONStringArray(raw string) []string {
	var out []string
	_ = json.Unmarshal([]byte(strings.TrimSpace(raw)), &out)
	return out
}

func parseJSONStringMap(raw string) map[string]string {
	out := map[string]string{}
	_ = json.Unmarshal([]byte(strings.TrimSpace(raw)), &out)
	return out
}

func sortedKeys(m map[string]interface{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
