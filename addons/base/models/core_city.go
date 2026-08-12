package models

import "sumeru/core/sdk"

type CoreCity struct {
	sdk.BaseModel
}

func (CoreCity) ModelName() string { return "core.city" }

func (CoreCity) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "City"},
		{Name: "state_id", Type: sdk.Many2One, Relation: "core.country.state", Index: true, String: "State"},
		{Name: "country_id", Type: sdk.Many2One, Relation: "core.country", Required: true, Index: true, String: "Country"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreCity{}, Module: "base"})
}
