package orm

// SysCron is a scheduled job definition (owned by automation module).
type SysCron struct {
	ID             int    `orm:"id"`
	Name           string `orm:"name"`
	Code           string `orm:"code"`
	EventName      string `orm:"event_name"`
	IntervalNumber int    `orm:"interval_number"`
	Active         bool   `orm:"active"`
	NextCall       string `orm:"next_call"`
	LastCall       string `orm:"last_call"`
}

func (SysCron) ModelName() string { return "sys.cron" }
func (SysCron) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "code", Type: Char, String: "Code"},
		{Name: "event_name", Type: Char, String: "Event Name"},
		{Name: "interval_number", Type: Integer, DefaultVal: 60, String: "Interval (minutes)"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
		{Name: "next_call", Type: DateTime, String: "Next Call"},
		{Name: "last_call", Type: DateTime, String: "Last Call"},
	}
}

// SysWorkflowTransition defines an allowed state change.
type SysWorkflowTransition struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`
	Model     string `orm:"model"`
	FromState string `orm:"from_state"`
	ToState   string `orm:"to_state"`
	GroupID   int    `orm:"group_id"`
	Active    bool   `orm:"active"`
}

func (SysWorkflowTransition) ModelName() string { return "sys.workflow.transition" }
func (SysWorkflowTransition) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "model", Type: Char, Required: true, Index: true, String: "Model"},
		{Name: "from_state", Type: Char, String: "From"},
		{Name: "to_state", Type: Char, Required: true, String: "To"},
		{Name: "group_id", Type: Many2One, Relation: "core.group", String: "Required Group"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
	}
}

// SysServerAction is a metadata action triggered by events or buttons.
type SysServerAction struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`
	Model     string `orm:"model"`
	EventName string `orm:"event_name"`
	Code      string `orm:"code"`
	Active    bool   `orm:"active"`
}

func (SysServerAction) ModelName() string { return "sys.server_action" }
func (SysServerAction) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "model", Type: Char, Index: true, String: "Model"},
		{Name: "event_name", Type: Char, Index: true, String: "On Event"},
		{Name: "code", Type: Text, String: "Code / Notes"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	const mod = "automation"
	RegisterModelWithModule(SysCron{}, mod)
	RegisterModelWithModule(SysWorkflowTransition{}, mod)
	RegisterModelWithModule(SysServerAction{}, mod)
}
