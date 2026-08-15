package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/report"
)

// RenderReportToolbar emits download/bulk controls when view declares report capabilities.
func RenderReportToolbar(caps report.Capabilities, model string, actionID int, menuID string, recordID int, fields []parser.Field, csrfToken string) string {
	if !caps.HasDownload() && !caps.BulkUpload {
		return ""
	}
	var fieldMeta []map[string]string
	for _, f := range fields {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		label := FieldDisplayLabel(f)
		fieldMeta = append(fieldMeta, map[string]string{"name": name, "label": label})
	}
	metaJSON, _ := json.Marshal(fieldMeta)
	capsJSON, _ := json.Marshal(caps)

	var sb strings.Builder
	sb.WriteString(`<div class="sum-report-toolbar" data-report-exchange`)
	sb.WriteString(` data-model="` + template.HTMLEscapeString(model) + `"`)
	sb.WriteString(` data-action-id="` + fmt.Sprintf("%d", actionID) + `"`)
	sb.WriteString(` data-menu-id="` + template.HTMLEscapeString(menuID) + `"`)
	if recordID > 0 {
		sb.WriteString(` data-record-id="` + fmt.Sprintf("%d", recordID) + `"`)
	}
	sb.WriteString(` data-fields="` + template.HTMLEscapeString(string(metaJSON)) + `"`)
	sb.WriteString(` data-caps="` + template.HTMLEscapeString(string(capsJSON)) + `"`)
	sb.WriteString(` data-csrf="` + template.HTMLEscapeString(csrfToken) + `"`)
	sb.WriteString(`>`)
	sb.WriteString(`<div class="sum-report-dropdown">`)
	sb.WriteString(`<button type="button" class="sum-list-btn-report" aria-haspopup="true">Report</button>`)
	sb.WriteString(`<div class="sum-report-menu" hidden>`)
	if caps.HasDownload() {
		sb.WriteString(`<button type="button" class="sum-report-item" data-report-action="export-csv">Download CSV</button>`)
		sb.WriteString(`<button type="button" class="sum-report-item" data-report-action="export-pdf">Download PDF</button>`)
	}
	if caps.BulkUpload {
		sb.WriteString(`<button type="button" class="sum-report-item" data-report-action="bulk-template">Download import template</button>`)
		sb.WriteString(`<button type="button" class="sum-report-item" data-report-action="bulk-upload">Bulk upload CSV</button>`)
	}
	sb.WriteString(`</div></div>`)
	sb.WriteString(`<div class="sum-report-modal" hidden aria-hidden="true">`)
	sb.WriteString(`<div class="sum-report-modal-inner">`)
	sb.WriteString(`<h2 class="sum-report-modal-title">Report options</h2>`)
	sb.WriteString(`<div class="sum-report-field-list"></div>`)
	sb.WriteString(`<label class="sum-report-pdf-size">PDF page size `)
	sb.WriteString(`<select class="sum-report-page-size"><option value="a4">A4</option><option value="legal">Legal</option><option value="letter">Letter</option></select>`)
	sb.WriteString(`</label>`)
	sb.WriteString(`<label class="sum-report-import-mode">Import mode `)
	sb.WriteString(`<select class="sum-report-import-mode-select"><option value="create">Create only</option><option value="upsert">Create or update</option></select>`)
	sb.WriteString(`</label>`)
	sb.WriteString(`<input type="file" class="sum-report-file-input" accept=".csv,text/csv" hidden />`)
	sb.WriteString(`<div class="sum-report-modal-actions">`)
	sb.WriteString(`<button type="button" class="sum-report-run">Run</button>`)
	sb.WriteString(`<button type="button" class="sum-report-cancel">Cancel</button>`)
	sb.WriteString(`</div></div></div>`)
	sb.WriteString(`</div>`)
	return sb.String()
}
