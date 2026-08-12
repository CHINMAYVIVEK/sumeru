package models

import "sumeru/core/sdk"

type CoreLang struct {
	sdk.BaseModel
	Code    string `db:"code"`
	Name    string `db:"name"`
	Active  bool   `db:"active"`
	ISOCode string `db:"iso_code"`
}

func (CoreLang) ModelName() string { return "core.lang" }

func (CoreLang) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "code", Type: sdk.Char, Required: true, Unique: true, Index: true, String: "Code"},
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
		{Name: "iso_code", Type: sdk.Char, String: "ISO Code"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreLang{}, Module: "base"})
}
