package models

import "sumeru/core/sdk"

// SysAttachment stores file metadata (content as base64/text or path).
type SysAttachment struct {
	sdk.BaseModel
	Name       string `db:"name"`
	Model      string `db:"model"`
	ResID      int64  `db:"res_id"`
	Mimetype   string `db:"mimetype"`
	FileSize   int64  `db:"file_size"`
	Datas      string `db:"datas"`
	StoreFname string `db:"store_fname"`
	CreateDate string `db:"create_date"`
	CompanyID  int    `db:"company_id"`
}

func (SysAttachment) ModelName() string { return "sys.attachment" }

func (SysAttachment) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "model", Type: sdk.Char, Index: true, String: "Model"},
		{Name: "res_id", Type: sdk.Integer, Index: true, String: "Record"},
		{Name: "mimetype", Type: sdk.Char, String: "MIME Type"},
		{Name: "file_size", Type: sdk.Integer, String: "Size"},
		{Name: "datas", Type: sdk.Text, String: "Data"},
		{Name: "store_fname", Type: sdk.Char, String: "Stored Filename"},
		{Name: "create_date", Type: sdk.DateTime, String: "Created"},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company", String: "Company"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysAttachment{}, Module: "base"})
}
