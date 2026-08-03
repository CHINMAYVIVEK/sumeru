package models

import "sumeru/core/sdk"

// SysCron is a scheduled job definition.
type SysCron struct {
	sdk.BaseModel
	Name           string `db:"name"`
	Code           string `db:"code"`
	EventName      string `db:"event_name"`
	IntervalNumber int    `db:"interval_number"`
	Active         bool   `db:"active"`
	NextCall       string `db:"next_call"`
	LastCall       string `db:"last_call"`
}

func (SysCron) ModelName() string { return "sys.cron" }

func (SysCron) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "code", Type: sdk.Char, String: "Code"},
		{Name: "event_name", Type: sdk.Char, String: "Event Name"},
		{Name: "interval_number", Type: sdk.Integer, DefaultVal: 60, String: "Interval (minutes)"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
		{Name: "next_call", Type: sdk.DateTime, String: "Next Call"},
		{Name: "last_call", Type: sdk.DateTime, String: "Last Call"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysCron{}, Module: "automation"})
}
