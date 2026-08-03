package orm

// SysActionWindow defines window actions
type SysActionWindow struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`
	CoreModel string `orm:"core_model"`
	ViewMode  string `orm:"view_mode"`
	Domain    string `orm:"domain"`
	Context   string `orm:"context"`
	Help      string `orm:"help"`
}

func (a SysActionWindow) ModelName() string { return "sys.action.window" }
func (a SysActionWindow) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "core_model", Type: Char, Required: true},
		{Name: "view_mode", Type: Char},
		{Name: "domain", Type: Char},
		{Name: "context", Type: Char},
		{Name: "help", Type: Char},
	}
}

func init() {
	RegisterModelWithModule(SysActionWindow{}, "base")
}
