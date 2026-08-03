package models

import "sumeru/core/sdk"

// MailActivity is a scheduled follow-up on a record (platform companion to chatter).
type MailActivity struct {
	sdk.BaseModel
	Name       string `db:"name"`
	Model      string `db:"model"`
	ResID      int64  `db:"res_id"`
	UserID     int    `db:"user_id"`
	Summary    string `db:"summary"`
	DateDeadline string `db:"date_deadline"`
	State      string `db:"state"` // planned | done | cancelled
}

func (MailActivity) ModelName() string { return "mail.activity" }

func (MailActivity) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Activity"},
		{Name: "model", Type: sdk.Char, Required: true, Index: true, String: "Model"},
		{Name: "res_id", Type: sdk.Integer, Required: true, Index: true, String: "Record"},
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", String: "Assigned To"},
		{Name: "summary", Type: sdk.Text, String: "Summary"},
		{Name: "date_deadline", Type: sdk.Date, String: "Due Date"},
		{Name: "state", Type: sdk.Selection, DefaultVal: "planned", String: "State",
			Selection: [][]string{{"planned", "Planned"}, {"done", "Done"}, {"cancelled", "Cancelled"}}},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &MailActivity{}, Module: "mail"})
}
