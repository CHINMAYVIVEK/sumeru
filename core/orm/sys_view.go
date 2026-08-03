package orm

// SysView stores XML view definitions
type SysView struct {
	ID       int    `orm:"id"`
	Name     string `orm:"name"`
	Model    string `orm:"model"` // Linked model technical name
	Type     string `orm:"type"`  // form, tree, kanban
	Arch     string `orm:"arch"`  // XML content
	Priority int    `orm:"priority"`
	Active   bool   `orm:"active"`
}

func (v SysView) ModelName() string { return "sys.view" }
func (v SysView) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "type", Type: Char, Required: true},
		{Name: "arch", Type: Text},
		{Name: "priority", Type: Integer},
		{Name: "active", Type: Boolean},
	}
}

func init() {
	RegisterModelWithModule(SysView{}, "base")
}
