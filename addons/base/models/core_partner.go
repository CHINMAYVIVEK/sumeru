package models

import (
	"sumeru/core/base"
)

type Partner struct {
	base.BaseModel
}

func (p Partner) ModelName() string {
	return "core.partner"
}

func (p Partner) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Name", Required: true},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{Name: "street", Type: base.Char, String: "Street"},
		{Name: "city", Type: base.Char, String: "City"},
		{Name: "country_id", Type: base.Many2One, Relation: "core.country", String: "Country"},
		{Name: "comment", Type: base.Text, String: "Notes"},
		{Name: "is_company", Type: base.Boolean, String: "Is a Company"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &Partner{}})
}
