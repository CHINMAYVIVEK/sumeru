package models

import "sumeru/core/sdk"

// SysServerAction is a metadata action triggered by events or buttons.
type SysServerAction struct {
	sdk.BaseModel
	Name      string `db:"name"`
	Model     string `db:"model"`
	EventName string `db:"event_name"`
	Code      string `db:"code"`
	Active    bool   `db:"active"`
}

func (SysServerAction) ModelName() string { return "sys.server.action" }

func (SysServerAction) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "model", Type: sdk.Char, Index: true, String: "Model"},
		{Name: "event_name", Type: sdk.Char, Index: true, String: "On Event"},
		{Name: "code", Type: sdk.Text, String: "Code / Notes"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysServerAction{}, Module: "automation"})
}
