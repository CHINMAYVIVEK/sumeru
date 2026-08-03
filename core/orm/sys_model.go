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

// SysField represents sys.field (metadata about model fields)
type SysField struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`       // field name e.g. login
	ModelID   int    `orm:"model_id"`   // reference to sys.model
	CoreModel string `orm:"core_model"` // technical name of parent model (denormalized)
	FieldType string `orm:"field_type"` // Char, Integer, etc.
	Relation  string `orm:"relation"`   // For Many2One, etc.
	Label     string `orm:"label"`      // String label
	Required  bool   `orm:"required"`
	Readonly  bool   `orm:"readonly"`
	Index     bool   `orm:"index"`
}

func (f SysField) ModelName() string { return "sys.field" }
func (f SysField) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true},
		{Name: "model_id", Type: Many2One, Relation: "sys.model", Required: true},
		{Name: "core_model", Type: Char},
		{Name: "field_type", Type: Char},
		{Name: "relation", Type: Char},
		{Name: "label", Type: Char},
		{Name: "required", Type: Boolean},
		{Name: "readonly", Type: Boolean},
		{Name: "index", Type: Boolean},
	}
}

func init() {
	const kernel = "base"
	RegisterModelWithModule(SysModel{}, kernel)
	RegisterModelWithModule(SysField{}, kernel)
}
