package orm

// SysModelData maps XML IDs to database IDs
type SysModelData struct {
	ID     int    `orm:"id"`
	Module string `orm:"module"`
	Name   string `orm:"name"`
	Model  string `orm:"model"`
	CoreID int    `orm:"core_id"`
}

func (d SysModelData) ModelName() string { return "sys.model_data" }
func (d SysModelData) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "module", Type: Char, Required: true},
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "core_id", Type: Integer, Required: true},
	}
}

func init() {
	RegisterModelWithModule(SysModelData{}, "base")
}
