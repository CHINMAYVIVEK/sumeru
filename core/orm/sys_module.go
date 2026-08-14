package orm

// SysModule tracks installable addons.
type SysModule struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"` // technical name
	DisplayName string `orm:"display_name"`
	Author      string `orm:"author"`
	Version     string `orm:"version"`
	Description string `orm:"description"`
	Icon        string `orm:"icon"` // optional relative path to module icon image
	State       string `orm:"state"` // uninstalled | to_install | installed | to_upgrade | to_remove
	Application bool   `orm:"application"`
	Active      bool   `orm:"active"`
	LastError   string `orm:"last_error"`
}

func (m SysModule) ModelName() string { return "sys.module" }
func (m SysModule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "display_name", Type: Char},
		{Name: "author", Type: Char},
		{Name: "version", Type: Char},
		{Name: "description", Type: Text},
		{Name: "icon", Type: Char},
		{Name: "state", Type: Char, Required: true},
		{Name: "application", Type: Boolean, Required: true},
		{Name: "active", Type: Boolean, Required: true},
		{Name: "last_error", Type: Text},
	}
}

func init() {
	RegisterModelWithModule(SysModule{}, "base")
}
