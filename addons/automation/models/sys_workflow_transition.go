package models

import "sumeru/core/sdk"

// SysWorkflowTransition defines an allowed state change.
type SysWorkflowTransition struct {
	sdk.BaseModel
	Name      string `db:"name"`
	Model     string `db:"model"`
	FromState string `db:"from_state"`
	ToState   string `db:"to_state"`
	GroupID   int    `db:"group_id"`
	Active    bool   `db:"active"`
}

func (SysWorkflowTransition) ModelName() string { return "sys.workflow.transition" }

func (SysWorkflowTransition) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "model", Type: sdk.Char, Required: true, Index: true, String: "Model"},
		{Name: "from_state", Type: sdk.Char, String: "From"},
		{Name: "to_state", Type: sdk.Char, Required: true, String: "To"},
		{Name: "group_id", Type: sdk.Many2One, Relation: "core.group", String: "Required Group"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysWorkflowTransition{}, Module: "automation"})
}
