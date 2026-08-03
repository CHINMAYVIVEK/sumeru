package models

import "sumeru/core/base"

type CoreUser struct {
	base.BaseModel
	Login     string `db:"login"`
	Password  string `db:"password"` // bcrypt hash; never list in UI
	Name      string `db:"name"`
	Active    bool   `db:"active"`
	Email     string `db:"email"`
	Phone     string `db:"phone"`
	Mobile    string `db:"mobile"`
	CompanyID int    `db:"company_id"`
	Lang      string `db:"lang"`
	TZ        string `db:"tz"`
	Signature string `db:"signature"`
	UserType     string `db:"user_type"` // internal | portal | public
	TOTPSecret   string `db:"totp_secret"`
	TOTPEnabled  bool   `db:"totp_enabled"`
	PasswordMinLen int  `db:"password_min_len"`
	// company_ids: M2M via core.user.company.rel (join table; not a SQL column)
}

func (CoreUser) ModelName() string { return "core.user" }

func (CoreUser) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "login", Type: base.Char, Required: true, Unique: true, String: "Login", Index: true},
		{Name: "password", Type: base.Char, String: "Password"},
		{Name: "name", Type: base.Char, String: "Name"},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Work Phone"},
		{Name: "mobile", Type: base.Char, String: "Mobile"},
		{Name: "company_id", Type: base.Many2One, Relation: "core.company", String: "Company", Index: true},
		{Name: "company_ids", Type: base.Many2Many, Relation: "core.company", RelationTable: "core_user_company_rel", Column1: "user_id", Column2: "company_id", String: "Companies"},
		{Name: "lang", Type: base.Selection, String: "Language", DefaultVal: "en_US"},
		{Name: "tz", Type: base.Char, String: "Timezone"},
		{Name: "signature", Type: base.Text, String: "Email Signature"},
		{Name: "user_type", Type: base.Selection, String: "User Type", DefaultVal: "internal",
			Selection: [][]string{{"internal", "Internal"}, {"portal", "Portal"}, {"public", "Public"}}},
		{Name: "totp_secret", Type: base.Char, String: "TOTP Secret"},
		{Name: "totp_enabled", Type: base.Boolean, DefaultVal: false, String: "2FA Enabled"},
		{Name: "password_min_len", Type: base.Integer, DefaultVal: 8, String: "Min Password Length"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CoreUser{}, Module: "base"})
}
