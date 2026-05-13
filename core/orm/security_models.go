package orm

// res.groups — application / functional roles (Odoo-style).
type ResGroups struct {
	ID         int    `orm:"id"`
	Name       string `orm:"name"`
	Category   string `orm:"category"`
	ParentID   int    `orm:"parent_id"`
	Sequence   int    `orm:"sequence"`
	Comment    string `orm:"comment"`
}

func (ResGroups) ModelName() string { return "res.groups" }
func (ResGroups) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "category", Type: Char},
		{Name: "parent_id", Type: Many2One},
		{Name: "sequence", Type: Integer},
		{Name: "comment", Type: Text},
	}
}

// res.groups.implied — group_id inherits implied_group_id.
type ResGroupsImplied struct {
	ID              int `orm:"id"`
	GroupID         int `orm:"group_id"`
	ImpliedGroupID int `orm:"implied_group_id"`
}

func (ResGroupsImplied) ModelName() string { return "res.groups.implied" }
func (ResGroupsImplied) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "group_id", Type: Many2One, Index: true},
		{Name: "implied_group_id", Type: Many2One, Index: true},
	}
}

// res.groups.user.rel — M2M between res.users and res.groups.
type ResGroupsUserRel struct {
	ID      int `orm:"id"`
	UserID  int `orm:"user_id"`
	GroupID int `orm:"group_id"`
}

func (ResGroupsUserRel) ModelName() string { return "res.groups.user.rel" }
func (ResGroupsUserRel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "user_id", Type: Many2One, Index: true},
		{Name: "group_id", Type: Many2One, Index: true},
	}
}

// ir.model.access — model-level CRUD permissions per group (or global when group_id is null).
type IrModelAccess struct {
	ID         int    `orm:"id"`
	Name       string `orm:"name"`
	Model      string `orm:"model"`
	GroupID    int    `orm:"group_id"` // 0 = treated as NULL in DB layer where needed
	PermRead   bool   `orm:"perm_read"`
	PermWrite  bool   `orm:"perm_write"`
	PermCreate bool   `orm:"perm_create"`
	PermUnlink bool   `orm:"perm_unlink"`
}

func (IrModelAccess) ModelName() string { return "ir.model.access" }
func (IrModelAccess) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true, Index: true},
		{Name: "group_id", Type: Many2One},
		{Name: "perm_read", Type: Boolean, Required: true, DefaultVal: false},
		{Name: "perm_write", Type: Boolean, Required: true, DefaultVal: false},
		{Name: "perm_create", Type: Boolean, Required: true, DefaultVal: false},
		{Name: "perm_unlink", Type: Boolean, Required: true, DefaultVal: false},
	}
}

// ir.rule — record-level domain rules.
type IrRule struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"`
	Model       string `orm:"model"`
	DomainForce string `orm:"domain_force"` // JSON array of triples, e.g. [["active","=",true]]
	Active      bool   `orm:"active"`
	PermRead    bool   `orm:"perm_read"`
	PermWrite   bool   `orm:"perm_write"`
	PermCreate  bool   `orm:"perm_create"`
	PermUnlink  bool   `orm:"perm_unlink"`
}

func (IrRule) ModelName() string { return "ir.rule" }
func (IrRule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true, Index: true},
		{Name: "domain_force", Type: Text},
		{Name: "active", Type: Boolean, Required: true, DefaultVal: true},
		{Name: "perm_read", Type: Boolean, Required: true, DefaultVal: true},
		{Name: "perm_write", Type: Boolean, Required: true, DefaultVal: true},
		{Name: "perm_create", Type: Boolean, Required: true, DefaultVal: true},
		{Name: "perm_unlink", Type: Boolean, Required: true, DefaultVal: true},
	}
}

// ir.rule.group.rel — rule applies only when user has one of these groups; empty = global rule.
type IrRuleGroupRel struct {
	ID      int `orm:"id"`
	RuleID  int `orm:"rule_id"`
	GroupID int `orm:"group_id"`
}

func (IrRuleGroupRel) ModelName() string { return "ir.rule.group.rel" }
func (IrRuleGroupRel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "rule_id", Type: Many2One, Index: true},
		{Name: "group_id", Type: Many2One, Index: true},
	}
}

// ir.session — server-side web sessions.
type IrSession struct {
	ID        int    `orm:"id"`
	Sid       string `orm:"sid"`
	UserID    int    `orm:"user_id"`
	ExpiresAt string `orm:"expires_at"` // ISO / timestamptz string from DB
}

func (IrSession) ModelName() string { return "ir.session" }
func (IrSession) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "sid", Type: Char, Required: true, Unique: true, Index: true},
		{Name: "user_id", Type: Many2One, Required: true, Index: true},
		{Name: "expires_at", Type: DateTime, Required: true},
	}
}

func init() {
	RegisterModel(ResGroups{})
	RegisterModel(ResGroupsImplied{})
	RegisterModel(ResGroupsUserRel{})
	RegisterModel(IrModelAccess{})
	RegisterModel(IrRule{})
	RegisterModel(IrRuleGroupRel{})
	RegisterModel(IrSession{})
}
