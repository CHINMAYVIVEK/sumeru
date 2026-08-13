package models

import "sumeru/core/sdk"

type SysConfigParameter struct {
	sdk.BaseModel
	Key   string `db:"key"`
	Value string `db:"value"`
}

func (SysConfigParameter) ModelName() string { return "sys.config.parameter" }

func (SysConfigParameter) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "key", Type: sdk.Char, Required: true, Unique: true, Index: true, String: "Key"},
		{Name: "value", Type: sdk.Text, String: "Value"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysConfigParameter{}, Module: "base"})
}
