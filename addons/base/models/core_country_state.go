package models

import "sumeru/core/sdk"

type CoreCountryState struct {
	sdk.BaseModel
}

func (CoreCountryState) ModelName() string { return "core.country.state" }

func (CoreCountryState) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "State"},
		{Name: "code", Type: sdk.Char, String: "Code"},
		{Name: "country_id", Type: sdk.Many2One, Relation: "core.country", Required: true, Index: true, String: "Country"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreCountryState{}, Module: "base"})
}
