package orm

// ResGroups represents Odoo-style res.groups
type ResGroups struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`
	Category  string `orm:"category"`
	Sequence  int    `orm:"sequence"`
}

func (g ResGroups) ModelName() string { return "res.groups" }
func (g ResGroups) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "category", Type: Char},
		{Name: "sequence", Type: Integer},
	}
}

// IrModelAccess represents Odoo-style ir.model.access
type IrModelAccess struct {
	ID         int    `orm:"id"`
	Name       string `orm:"name"`
	Model      string `orm:"model"` // Technical name e.g. res.partner
	GroupID    int    `orm:"group_id"`
	PermRead   bool   `orm:"perm_read"`
	PermWrite  bool   `orm:"perm_write"`
	PermCreate bool   `orm:"perm_create"`
	PermUnlink bool   `orm:"perm_unlink"`
}

func (a IrModelAccess) ModelName() string { return "ir.model.access" }
func (a IrModelAccess) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "group_id", Type: Many2One, Relation: "res.groups"},
		{Name: "perm_read", Type: Boolean},
		{Name: "perm_write", Type: Boolean},
		{Name: "perm_create", Type: Boolean},
		{Name: "perm_unlink", Type: Boolean},
	}
}

// IrRule represents Odoo-style record rules (ir.rule)
type IrRule struct {
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

func (r IrRule) ModelName() string { return "ir.rule" }
func (r IrRule) Fields() []FieldDefinition {
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

// IrApprovalRule represents a stage transition approval mechanism
type IrApprovalRule struct {
	ID              int    `orm:"id"`
	Model           string `orm:"model"`
	GroupID         int    `orm:"group_id"`
	FromState       string `orm:"from_state"`
	ToState         string `orm:"to_state"`
	RequireApproval bool   `orm:"require_approval"`
}

func (r IrApprovalRule) ModelName() string { return "ir.approval.rule" }
func (r IrApprovalRule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "model", Type: Char, Required: true},
		{Name: "group_id", Type: Many2One, Relation: "res.groups", Required: true},
		{Name: "from_state", Type: Char},
		{Name: "to_state", Type: Char, Required: true},
		{Name: "require_approval", Type: Boolean},
	}
}

func init() {
	RegisterModel(ResGroups{})
	RegisterModel(IrModelAccess{})
	RegisterModel(IrRule{})
	RegisterModel(IrApprovalRule{})
}
