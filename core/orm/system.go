package orm

// ir_model stores information about models
type IrModel struct {
	ID        int       `orm:"id"`
	Name      string    `orm:"name"`       // Technical name: sale.order
	Model     string    `orm:"model"`      // Human name: Sales Order
	Info      string    `orm:"info"`
	Transient bool      `orm:"transient"`
}

func (m IrModel) ModelName() string { return "ir.model" }
func (m IrModel) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "info", Type: Char},
		{Name: "transient", Type: Boolean},
	}
}

// ir_ui_view stores XML view definitions
type IrUiView struct {
	ID        int       `orm:"id"`
	Name      string    `orm:"name"`
	Model     string    `orm:"model"`      // Linked model
	Type      string    `orm:"type"`       // form, tree, kanban
	Arch      string    `orm:"arch"`       // XML content
	Priority  int       `orm:"priority"`
}

func (v IrUiView) ModelName() string { return "ir.ui.view" }
func (v IrUiView) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "type", Type: Char, Required: true},
		{Name: "arch", Type: Text},
		{Name: "priority", Type: Integer},
	}
}

// ir_ui_menu stores menu hierarchy
type IrUiMenu struct {
	ID            int       `orm:"id"`
	Name          string    `orm:"name"`
	ParentID      int       `orm:"parent_id"`
	ActionID      int       `orm:"action_id"`
	Sequence      int       `orm:"sequence"`
	WebIcon       string    `orm:"web_icon"`
	Module        string    `orm:"module"`         // technical addon that owns this menu (for filtering by installed apps)
	AccessGroups  string    `orm:"access_groups"` // comma-separated XML ids (e.g. user.group_admin); empty = all users
}

func (m IrUiMenu) ModelName() string { return "ir.ui.menu" }
func (m IrUiMenu) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "parent_id", Type: Many2One},
		{Name: "action_id", Type: Integer},
		{Name: "sequence", Type: Integer},
		{Name: "web_icon", Type: Char},
		{Name: "module", Type: Char, Index: true},
		{Name: "access_groups", Type: Char},
	}
}

// ir.module tracks installable addons (registry metadata).
type IrModule struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"`         // technical name, matches manifest name / addon folder
	DisplayName string `orm:"display_name"` // human label in Apps
	Author      string `orm:"author"`
	Version     string `orm:"version"`
	Description string `orm:"description"`
	State       string `orm:"state"` // installed | uninstalled
	Application bool   `orm:"application"` // show in Apps dashboard
	Active      bool   `orm:"active"`      // false = deactivated (menus hidden, data kept)
}

func (m IrModule) ModelName() string { return "ir.module" }
func (m IrModule) Fields() []FieldDefinition {
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

// ir_actions_act_window defines what happens when a menu is clicked
type IrActionsActWindow struct {
	ID        int       `orm:"id"`
	Name      string    `orm:"name"`
	ResModel  string    `orm:"res_model"`
	ViewMode  string    `orm:"view_mode"`  // tree,form,kanban
	Domain    string    `orm:"domain"`
	Context   string    `orm:"context"`
	Help      string    `orm:"help"`
}

func (a IrActionsActWindow) ModelName() string { return "ir.actions.act_window" }
func (a IrActionsActWindow) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "res_model", Type: Char, Required: true},
		{Name: "view_mode", Type: Char},
		{Name: "domain", Type: Char},
		{Name: "context", Type: Char},
		{Name: "help", Type: Char},
	}
}

// IrSession stores web session tokens for authenticated users.
type IrSession struct {
	ID        int    `orm:"id"`
	Sid       string `orm:"sid"`
	UserID    int    `orm:"user_id"`
	ExpiresAt string `orm:"expires_at"`
}

func (s IrSession) ModelName() string { return "ir.session" }
func (s IrSession) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "sid", Type: Char, Required: true, Unique: true, Index: true},
		{Name: "user_id", Type: Many2One, Relation: "res.users", Required: true, Index: true},
		{Name: "expires_at", Type: DateTime, Required: true},
	}
}

// mail_message stores chatter and activity log lines (Odoo-style subset).
type MailMessage struct {
	ID         int       `orm:"id"`
	Model      string    `orm:"model"`       // technical model name e.g. res.company
	ResID      int64     `orm:"res_id"`      // row id in that model
	Body       string    `orm:"body"`
	Subtype    string    `orm:"subtype"`     // comment | notification | module
	Author     string    `orm:"author"`      // display name until res.users is wired
	CreateDate string    `orm:"create_date"` // set on insert (TIMESTAMPTZ)
	CompanyID  int64     `orm:"company_id"`  // optional FK res.company
}

func (m MailMessage) ModelName() string { return "mail.message" }
func (m MailMessage) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "model", Type: Char, Required: true, Index: true},
		{Name: "res_id", Type: Integer, Required: true},
		{Name: "body", Type: Text, Required: true},
		{Name: "subtype", Type: Char, Required: true},
		{Name: "author", Type: Char},
		{Name: "create_date", Type: DateTime, Required: true},
		{Name: "company_id", Type: Many2One},
	}
}

// ir_model_data maps XML IDs to database IDs
type IrModelData struct {
	ID      int    `orm:"id"`
	Module  string `orm:"module"`
	Name    string `xml:"name,attr"` // XML ID
	Model   string `orm:"model"`
	ResID   int    `orm:"res_id"`
}

func (d IrModelData) ModelName() string { return "ir.model.data" }
func (d IrModelData) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "module", Type: Char, Required: true},
		{Name: "name", Type: Char, Required: true, Unique: true},
		{Name: "model", Type: Char, Required: true},
		{Name: "res_id", Type: Integer, Required: true},
	}
}

func init() {
	RegisterModel(IrModel{})
	RegisterModel(IrUiView{})
	RegisterModel(IrUiMenu{})
	RegisterModel(IrActionsActWindow{})
	RegisterModel(IrModelData{})
	RegisterModel(IrModule{})
	RegisterModel(MailMessage{})
	RegisterModel(IrSession{})
}
