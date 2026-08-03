package orm

// SysMenu stores menu hierarchy
type SysMenu struct {
	ID           int    `orm:"id"`
	Name         string `orm:"name"`
	ParentID     int    `orm:"parent_id"`
	ActionID     int    `orm:"action_id"`
	Action       string `orm:"action"` // String ref for XML loading: model,id
	Sequence     int    `orm:"sequence"`
	WebIcon      string `orm:"web_icon"`
	Module       string `orm:"module"`        // technical addon that owns this menu
	AccessGroups string `orm:"access_groups"` // comma-separated XML ids
}

func (m SysMenu) ModelName() string { return "sys.menu" }
func (m SysMenu) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "parent_id", Type: Many2One, Relation: "sys.menu"},
		{Name: "action_id", Type: Integer},
		{Name: "action", Type: Char},
		{Name: "sequence", Type: Integer},
		{Name: "web_icon", Type: Char},
		{Name: "module", Type: Char, Index: true},
		{Name: "access_groups", Type: Char},
	}
}

func init() {
	RegisterModelWithModule(SysMenu{}, "base")
}
