package models

import "sumeru/core/base"

// CoreGroup represents core.group (formerly res.groups)
type CoreGroup struct {
	base.BaseModel
	Name     string `db:"name"`
	Category string `db:"category"`
	Sequence int    `db:"sequence"`
}

func (CoreGroup) ModelName() string { return "core.group" }
func (CoreGroup) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, Unique: true, String: "Name"},
		{Name: "category", Type: base.Char, String: "Category"},
		{Name: "sequence", Type: base.Integer, String: "Sequence"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CoreGroup{}, Module: "base"})
}
