package orm

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

func init() {
	RegisterModelWithModule(SysSession{}, "base")
}
