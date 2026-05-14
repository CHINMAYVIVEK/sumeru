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

// SysModule tracks installable addons.
type SysModule struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"` // technical name
	DisplayName string `orm:"display_name"`
	Author      string `orm:"author"`
	Version     string `orm:"version"`
	Description string `orm:"description"`
	State       string `orm:"state"` // installed | uninstalled
	Application bool   `orm:"application"`
	Active      bool   `orm:"active"`
}

func (m SysModule) ModelName() string { return "sys.module" }
func (m SysModule) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "display_name", Type: Char},
		{Name: "author", Type: Char},
		{Name: "version", Type: Char},
		{Name: "description", Type: Text},
		{Name: "state", Type: Char, Required: true},
		{Name: "application", Type: Boolean, Required: true},
		{Name: "active", Type: Boolean, Required: true},
	}
}

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

// SysSession stores web sessions
type SysSession struct {
	ID        int    `orm:"id"`
	Sid       string `orm:"sid"`
	UserID    int    `orm:"user_id"`
	ExpiresAt string `orm:"expires_at"`
}

func (s SysSession) ModelName() string { return "sys.session" }
func (s SysSession) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "sid", Type: Char, Required: true, Unique: true, Index: true},
		{Name: "user_id", Type: Many2One, Relation: "core.user", Required: true, Index: true},
		{Name: "expires_at", Type: DateTime, Required: true},
	}
}

// AppLog stores application and module lifecycle audit lines (install, update, etc.).
// Internal user-to-user chatter uses mail.message only.
type AppLog struct {
	ID         int    `orm:"id"`
	ModuleName string `orm:"module_name"`
	Action     string `orm:"action"`
	Detail     string `orm:"detail"`
	Author     string `orm:"author"`
	CreateDate string `orm:"create_date"`
}

func (AppLog) ModelName() string { return "app.log" }
func (AppLog) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "module_name", Type: Char, Required: true, Index: true},
		{Name: "action", Type: Char, Required: true},
		{Name: "detail", Type: Text},
		{Name: "author", Type: Char},
		{Name: "create_date", Type: DateTime, Required: true},
	}
}

// MailMessage stores chatter and activity log lines
type MailMessage struct {
	ID         int    `orm:"id"`
	Model      string `orm:"model"`   // technical model name
	CoreID     int64  `orm:"core_id"` // row id
	Body       string `orm:"body"`
	Subtype    string `orm:"subtype"`
	Author     string `orm:"author"`
	CreateDate string `orm:"create_date"`
	CompanyID  int64  `orm:"company_id"`
}

func (m MailMessage) ModelName() string { return "mail.message" }
func (m MailMessage) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "model", Type: Char, Required: true, Index: true},
		{Name: "core_id", Type: Integer, Required: true},
		{Name: "body", Type: Text, Required: true},
		{Name: "subtype", Type: Char, Required: true},
		{Name: "author", Type: Char},
		{Name: "create_date", Type: DateTime, Required: true},
		{Name: "company_id", Type: Many2One, Relation: "core.company"},
	}
}

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
	const kernel = "base"
	RegisterModelWithModule(SysModel{}, kernel)
	RegisterModelWithModule(SysField{}, kernel)
	RegisterModelWithModule(SysView{}, kernel)
	RegisterModelWithModule(SysMenu{}, kernel)
	RegisterModelWithModule(SysActionWindow{}, kernel)
	RegisterModelWithModule(SysModelData{}, kernel)
	RegisterModelWithModule(SysModule{}, kernel)
	RegisterModelWithModule(AppLog{}, kernel)
	RegisterModelWithModule(MailMessage{}, kernel)
	RegisterModelWithModule(SysSession{}, kernel)
}
