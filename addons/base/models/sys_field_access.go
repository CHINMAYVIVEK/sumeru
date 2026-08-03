package models

import "sumeru/core/sdk"

// SysFieldAccess is optional field-level security (read/write per group).
type SysFieldAccess struct {
	sdk.BaseModel
	Name      string `db:"name"`
	Model     string `db:"model"`
	FieldName string `db:"field_name"`
	GroupID   int    `db:"group_id"`
	PermRead  bool   `db:"perm_read"`
	PermWrite bool   `db:"perm_write"`
}

func (SysFieldAccess) ModelName() string { return "sys.field_access" }

func (SysFieldAccess) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, Unique: true, String: "Name"},
		{Name: "model", Type: sdk.Char, Required: true, Index: true, String: "Model"},
		{Name: "field_name", Type: sdk.Char, Required: true, String: "Field"},
		{Name: "group_id", Type: sdk.Many2One, Relation: "core.group", String: "Group"},
		{Name: "perm_read", Type: sdk.Boolean, DefaultVal: true, String: "Read"},
		{Name: "perm_write", Type: sdk.Boolean, DefaultVal: true, String: "Write"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysFieldAccess{}, Module: "base"})
}
