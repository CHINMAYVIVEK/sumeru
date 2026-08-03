package models

import "sumeru/core/sdk"

// SysSequence generates monotonic document numbers.
type SysSequence struct {
	sdk.BaseModel
	Name       string `db:"name"`
	Code       string `db:"code"`
	Prefix     string `db:"prefix"`
	Suffix     string `db:"suffix"`
	Padding    int    `db:"padding"`
	NumberNext int    `db:"number_next"`
	Active     bool   `db:"active"`
}

func (SysSequence) ModelName() string { return "sys.sequence" }

func (SysSequence) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "code", Type: sdk.Char, Required: true, Unique: true, Index: true, String: "Code"},
		{Name: "prefix", Type: sdk.Char, String: "Prefix"},
		{Name: "suffix", Type: sdk.Char, String: "Suffix"},
		{Name: "padding", Type: sdk.Integer, DefaultVal: 5, String: "Padding"},
		{Name: "number_next", Type: sdk.Integer, DefaultVal: 1, String: "Next Number"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysSequence{}, Module: "base"})
}
