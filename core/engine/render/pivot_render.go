package render

import (
	"context"
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

// BuildPivotData aggregates list rows for a pivot view definition.
func BuildPivotData(view *parser.View, rows []map[string]interface{}) *PivotData {
	if view == nil || len(view.Field) == 0 {
		return nil
	}
	var rowFields, colFields, measureFields []parser.Field
	for _, f := range view.Field {
		switch strings.ToLower(strings.TrimSpace(f.PivotType)) {
		case "row":
			rowFields = append(rowFields, f)
		case "col", "column":
			colFields = append(colFields, f)
		case "measure":
			measureFields = append(measureFields, f)
		}
	}
	if len(measureFields) == 0 {
		return nil
	}
	measure := measureFields[0]
	measureLabel := FieldDisplayLabel(measure)

	values := map[string]map[string]float64{}
	rowSet := map[string]struct{}{}
	colSet := map[string]struct{}{}

	for _, row := range rows {
		rowKey := pivotKey(row, rowFields)
		colKey := pivotKey(row, colFields)
		if colKey == "" {
			colKey = "Total"
		}
		rowSet[rowKey] = struct{}{}
		colSet[colKey] = struct{}{}
		if values[rowKey] == nil {
			values[rowKey] = map[string]float64{}
		}
		values[rowKey][colKey] += pivotMeasure(row, measure.Name)
	}

	out := &PivotData{
		RowLabels:    pivotSortedKeys(rowSet),
		ColLabels:    pivotSortedKeys(colSet),
		Values:       values,
		MeasureLabel: measureLabel,
	}
	if len(out.RowLabels) == 0 {
		out.RowLabels = []string{"Total"}
	}
	return out
}

func pivotKey(row map[string]interface{}, fields []parser.Field) string {
	if len(fields) == 0 {
		return "Total"
	}
	parts := make([]string, 0, len(fields))
	for _, f := range fields {
		parts = append(parts, pivotCellLabel(row, f.Name))
	}
	return strings.Join(parts, " / ")
}

func pivotCellLabel(row map[string]interface{}, fieldName string) string {
	if v, ok := row[fieldName+"_name"]; ok && orm.AsString(v) != "" {
		return orm.AsString(v)
	}
	v := row[fieldName]
	if v == nil {
		return "—"
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func pivotMeasure(row map[string]interface{}, fieldName string) float64 {
	v := row[fieldName]
	switch n := v.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(n), 64)
		return f
	default:
		f, _ := strconv.ParseFloat(strings.TrimSpace(fmt.Sprint(v)), 64)
		return f
	}
}

func pivotSortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// RenderPivot builds HTML for a pivot workspace view.
func RenderPivot(_ context.Context, view *parser.View, data *PivotData) string {
	if data == nil || len(data.ColLabels) == 0 {
		return pivotPlaceholder()
	}
	var sb strings.Builder
	title := HumanViewBreadcrumb(view.Model, "pivot")
	sb.WriteString(`<div class="sum-pivot-view">`)
	sb.WriteString(`<h1 class="sum-pivot-title">` + template.HTMLEscapeString(title) + `</h1>`)
	sb.WriteString(`<div class="sum-pivot-table-wrap"><table class="sum-pivot-table">`)
	sb.WriteString(`<thead><tr><th></th>`)
	for _, col := range data.ColLabels {
		sb.WriteString(`<th>` + template.HTMLEscapeString(col) + `</th>`)
	}
	sb.WriteString(`</tr></thead><tbody>`)
	for _, rowLabel := range data.RowLabels {
		sb.WriteString(`<tr><th scope="row">` + template.HTMLEscapeString(rowLabel) + `</th>`)
		rowVals := data.Values[rowLabel]
		for _, col := range data.ColLabels {
			val := rowVals[col]
			sb.WriteString(`<td>` + template.HTMLEscapeString(formatPivotNumber(val)) + `</td>`)
		}
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table></div>`)
	if data.MeasureLabel != "" {
		sb.WriteString(`<p class="sum-pivot-measure">Measure: ` + template.HTMLEscapeString(data.MeasureLabel) + `</p>`)
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func formatPivotNumber(v float64) string {
	if v == float64(int64(v)) {
		return fmt.Sprintf("%d", int64(v))
	}
	return fmt.Sprintf("%.2f", v)
}

func pivotPlaceholder() string {
	var sb strings.Builder
	sb.WriteString(`<div class="sum-pivot-placeholder">`)
	sb.WriteString(`<svg class="sum-pivot-placeholder-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>`)
	sb.WriteString(`<h3 class="sum-pivot-placeholder-title">Pivot view</h3>`)
	sb.WriteString(`<p class="sum-pivot-placeholder-text">Add pivot fields with type="row", type="col", and type="measure" in the view arch.</p>`)
	sb.WriteString(`</div>`)
	return sb.String()
}
