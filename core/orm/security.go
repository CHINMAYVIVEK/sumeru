package orm

// SysAccess represents sys.access (formerly sys.model.access)
type SysAccess struct {
	ID         int    `orm:"id"`
	Name       string `orm:"name"`
	Model      string `orm:"model"` // Technical name e.g. core.partner
	GroupID    int    `orm:"group_id"`
	PermRead   bool   `orm:"perm_read"`
	PermWrite  bool   `orm:"perm_write"`
	PermCreate bool   `orm:"perm_create"`
	PermUnlink bool   `orm:"perm_unlink"`
}

func (a SysAccess) ModelName() string { return "sys.access" }
func (a SysAccess) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "group_id", Type: Many2One, Relation: "core.group"},
		{Name: "perm_read", Type: Boolean},
		{Name: "perm_write", Type: Boolean},
		{Name: "perm_create", Type: Boolean},
		{Name: "perm_unlink", Type: Boolean},
	}
}

// SysRule represents sys.rule (record rules)
type SysRule struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"`
	Model       string `orm:"model"`
	DomainForce string `orm:"domain_force"`
	Active      bool   `orm:"active"`
	PermRead    bool   `orm:"perm_read"`
	PermWrite   bool   `orm:"perm_write"`
	PermCreate  bool   `orm:"perm_create"`
	PermUnlink  bool   `orm:"perm_unlink"`
}

func (r SysRule) ModelName() string { return "sys.rule" }
func (r SysRule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "domain_force", Type: Text},
		{Name: "active", Type: Boolean},
		{Name: "perm_read", Type: Boolean},
		{Name: "perm_write", Type: Boolean},
		{Name: "perm_create", Type: Boolean},
		{Name: "perm_unlink", Type: Boolean},
	}
}

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
	const kernel = "base"
	RegisterModelWithModule(SysAccess{}, kernel)
	RegisterModelWithModule(SysRule{}, kernel)
	RegisterModelWithModule(SysApprovalRule{}, kernel)
}
