package orm

// SysModuleCategory classifies core.group rows for display (name + sequence).
type SysModuleCategory struct {
	ID       int    `orm:"id"`
	Name     string `orm:"name"`
	Sequence int    `orm:"sequence"`
}

func (SysModuleCategory) ModelName() string { return "sys.module.category" }
func (SysModuleCategory) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true, String: "Name"},
		{Name: "sequence", Type: Integer, String: "Sequence"},
	}
}

func init() {
	RegisterModelWithModule(SysModuleCategory{}, "base")
}
