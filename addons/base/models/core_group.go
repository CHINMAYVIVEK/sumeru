package models

import "sumeru/core/sdk"

type CoreGroup struct {
	sdk.BaseModel
	Name        string `db:"name"`
	CategoryID  int    `db:"category_id"`
	Sequence    int    `db:"sequence"`
}

func (CoreGroup) ModelName() string { return "core.group" }
func (CoreGroup) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, Unique: true, String: "Name"},
		{Name: "category_id", Type: sdk.Many2One, Relation: "sys.module.category", String: "Application", Index: true},
		{Name: "sequence", Type: sdk.Integer, String: "Sequence"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreGroup{}, Module: "base"})
}
