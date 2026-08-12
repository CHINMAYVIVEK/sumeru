package orm

// SysAccess is per-model CRUD permission tied to an optional core.group.
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

func init() {
	RegisterModelWithModule(SysAccess{}, "base")
}
