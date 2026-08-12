package models

import "sumeru/core/sdk"

type CoreUserLog struct {
	sdk.BaseModel
	UserID     int    `db:"user_id"`
	CreateDate string `db:"create_date"`
	IP         string `db:"ip"`
	Result     string `db:"result"` // success | failure
}

func (CoreUserLog) ModelName() string { return "core.user.log" }

func (CoreUserLog) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", Index: true, String: "User"},
		{Name: "create_date", Type: sdk.DateTime, Required: true, String: "When"},
		{Name: "ip", Type: sdk.Char, String: "IP"},
		{Name: "result", Type: sdk.Char, Required: true, DefaultVal: "success", String: "Result"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreUserLog{}, Module: "base"})
}
