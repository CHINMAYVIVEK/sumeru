package models

import (
	"sumeru/core/sdk"
)

type Partner struct {
	sdk.BaseModel
}

func (p Partner) ModelName() string {
	return "core.partner"
}

func (p Partner) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, String: "Name", Required: true},
		{Name: "image", Type: sdk.Text, String: "Image"},
		{Name: "email", Type: sdk.Char, String: "Email"},
		{Name: "phone", Type: sdk.Char, String: "Phone"},
		{Name: "street", Type: sdk.Char, String: "Street"},
		{Name: "city", Type: sdk.Char, String: "City"},
		{Name: "country_id", Type: sdk.Many2One, Relation: "core.country", String: "Country"},
		{Name: "comment", Type: sdk.Text, String: "Notes"},
		{Name: "is_company", Type: sdk.Boolean, String: "Is a Company"},
		{Name: "active", Type: sdk.Boolean, String: "Active", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &Partner{}, Module: "base"})
}
