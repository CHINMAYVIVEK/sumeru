package models

import "sumeru/core/base"

// MailActivity is a scheduled follow-up on a record (platform companion to chatter).
type MailActivity struct {
	base.BaseModel
	Name       string `db:"name"`
	Model      string `db:"model"`
	ResID      int64  `db:"res_id"`
	UserID     int    `db:"user_id"`
	Summary    string `db:"summary"`
	DateDeadline string `db:"date_deadline"`
	State      string `db:"state"` // planned | done | cancelled
}

func (MailActivity) ModelName() string { return "mail.activity" }

func (MailActivity) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Activity"},
		{Name: "model", Type: base.Char, Required: true, Index: true, String: "Model"},
		{Name: "res_id", Type: base.Integer, Required: true, Index: true, String: "Record"},
		{Name: "user_id", Type: base.Many2One, Relation: "core.user", String: "Assigned To"},
		{Name: "summary", Type: base.Text, String: "Summary"},
		{Name: "date_deadline", Type: base.Date, String: "Due Date"},
		{Name: "state", Type: base.Selection, DefaultVal: "planned", String: "State",
			Selection: [][]string{{"planned", "Planned"}, {"done", "Done"}, {"cancelled", "Cancelled"}}},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &MailActivity{}, Module: "mail"})
}
