package models

import "sumeru/core/base"

// CoreGroup is the core.group access-group model (users belong via core.group.user.rel).
type CoreGroup struct {
	base.BaseModel
	Name        string `db:"name"`
	CategoryID  int    `db:"category_id"`
	Sequence    int    `db:"sequence"`
}

func (CoreGroup) ModelName() string { return "core.group" }
func (CoreGroup) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, Unique: true, String: "Name"},
		{Name: "category_id", Type: base.Many2One, Relation: "sys.module.category", String: "Application", Index: true},
		{Name: "sequence", Type: base.Integer, String: "Sequence"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CoreGroup{}, Module: "base"})
}
