package orm

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

func init() {
	RegisterModelWithModule(SysRule{}, "base")
}
