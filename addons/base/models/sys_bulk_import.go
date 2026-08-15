package models

import "sumeru/core/sdk"

// SysBulkImport is a transient batch row for CSV bulk upload mapping and confirm.
type SysBulkImport struct {
	sdk.BaseModel
}

func (SysBulkImport) ModelName() string { return "sys.bulk.import" }

func (SysBulkImport) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "target_model", Type: sdk.Char, Required: true, String: "Target Model"},
		{Name: "import_mode", Type: sdk.Char, String: "Import Mode"},
		{Name: "attachment_id", Type: sdk.Many2One, Relation: "sys.attachment", String: "Staged File"},
		{Name: "selected_fields", Type: sdk.Text, String: "Selected Fields"},
		{Name: "csv_headers", Type: sdk.Text, String: "CSV Headers"},
		{Name: "column_mapping", Type: sdk.Text, String: "Column Mapping"},
		{Name: "preview_json", Type: sdk.Text, String: "Preview"},
		{Name: "next_url", Type: sdk.Char, String: "Return URL"},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "User"},
		{Name: "action_id", Type: sdk.Integer, String: "Source Action"},
		{Name: "state", Type: sdk.Char, String: "State"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysBulkImport{}, Module: "base"})
}
