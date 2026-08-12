package orm

// SysModel stores information about models
type SysModel struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"`   // Technical name: core.user
	Model       string `orm:"model"`  // Human name: Users
	Module      string `orm:"module"` // sys.module technical name of declaring addon (kernel → base)
	Description string `orm:"description"`
	Transient   bool   `orm:"transient"`
}

func (m SysModel) ModelName() string { return "sys.model" }
func (m SysModel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "module", Type: Char, String: "Declaring module", Index: true},
		{Name: "description", Type: Text},
		{Name: "transient", Type: Boolean},
	}
}

func init() {
	RegisterModelWithModule(SysModel{}, "base")
}
