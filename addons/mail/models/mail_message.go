package models

import "sumeru/core/sdk"

// MailMessage stores chatter and activity log lines.
type MailMessage struct {
	sdk.BaseModel
	Model      string `db:"model"`
	CoreID     int64  `db:"core_id"`
	Body       string `db:"body"`
	Subtype    string `db:"subtype"`
	Author     string `db:"author"`
	CreateDate string `db:"create_date"`
	CompanyID  int64  `db:"company_id"`
}

func (MailMessage) ModelName() string { return "mail.message" }

func (MailMessage) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "model", Type: sdk.Char, Required: true, Index: true},
		{Name: "core_id", Type: sdk.Integer, Required: true},
		{Name: "body", Type: sdk.Text, Required: true},
		{Name: "subtype", Type: sdk.Char, Required: true},
		{Name: "author", Type: sdk.Char},
		{Name: "create_date", Type: sdk.DateTime, Required: true},
		{Name: "company_id", Type: sdk.Many2One, Relation: "core.company"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &MailMessage{}, Module: "mail"})
}
