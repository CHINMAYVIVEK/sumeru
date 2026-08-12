package models

import "sumeru/core/sdk"

type SysAudit struct {
	sdk.BaseModel
	UserID     int    `db:"user_id"`
	Action     string `db:"action"` // create | write | unlink | login_fail | access_deny
	Model      string `db:"model"`
	ResID      int64  `db:"res_id"`
	BeforeJSON string `db:"before_json"`
	AfterJSON  string `db:"after_json"`
	Detail     string `db:"detail"`
	CreateDate string `db:"create_date"`
}

func (SysAudit) ModelName() string { return "sys.audit" }

func (SysAudit) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", Index: true, String: "User"},
		{Name: "action", Type: sdk.Char, Required: true, Index: true, String: "Action"},
		{Name: "model", Type: sdk.Char, Index: true, String: "Model"},
		{Name: "res_id", Type: sdk.Integer, String: "Record"},
		{Name: "before_json", Type: sdk.Text, String: "Before"},
		{Name: "after_json", Type: sdk.Text, String: "After"},
		{Name: "detail", Type: sdk.Text, String: "Detail"},
		{Name: "create_date", Type: sdk.DateTime, Required: true, Index: true, String: "When"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysAudit{}, Module: "base"})
}
