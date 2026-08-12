package orm

// SysApprovalRule represents stage transition approval
type SysApprovalRule struct {
	ID              int    `orm:"id"`
	Model           string `orm:"model"`
	GroupID         int    `orm:"group_id"`
	FromState       string `orm:"from_state"`
	ToState         string `orm:"to_state"`
	RequireApproval bool   `orm:"require_approval"`
}

func (r SysApprovalRule) ModelName() string { return "sys.approval_rule" }
func (r SysApprovalRule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "model", Type: Char, Required: true},
		{Name: "group_id", Type: Many2One, Relation: "core.group", Required: true},
		{Name: "from_state", Type: Char},
		{Name: "to_state", Type: Char, Required: true},
		{Name: "require_approval", Type: Boolean},
	}
}

func init() {
	RegisterModelWithModule(SysApprovalRule{}, "base")
}
