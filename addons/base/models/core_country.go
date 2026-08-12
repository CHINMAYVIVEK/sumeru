package models

import "sumeru/core/sdk"

// CoreCountry is a geographic country with calling code.
type CoreCountry struct {
	sdk.BaseModel
}

func (CoreCountry) ModelName() string { return "core.country" }

func (CoreCountry) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Country"},
		{Name: "code", Type: sdk.Char, Required: true, Unique: true, Index: true, String: "Code"},
		{Name: "phone_code", Type: sdk.Char, String: "Phone Code"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreCountry{}, Module: "base"})
}
