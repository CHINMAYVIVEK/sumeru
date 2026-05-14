package models

import "sumeru/core/base"

type ResUsers struct {
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
	UserType  string `db:"user_type"` // internal | portal | public
}

func (ResUsers) ModelName() string { return "res.users" }

func (ResUsers) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "login", Type: base.Char, Required: true, Unique: true, String: "Login", Index: true},
		{Name: "password", Type: base.Char, String: "Password"},
		{Name: "name", Type: base.Char, String: "Name"},
		{Name: "active", Type: base.Boolean, String: "Active", DefaultVal: true},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Work Phone"},
		{Name: "mobile", Type: base.Char, String: "Mobile"},
		{Name: "company_id", Type: base.Many2One, Relation: "res.company", String: "Company", Index: true},
		{Name: "lang", Type: base.Selection, String: "Language", DefaultVal: "en_US"},
		{Name: "tz", Type: base.Char, String: "Timezone"},
		{Name: "signature", Type: base.Text, String: "Email Signature"},
		{Name: "user_type", Type: base.Selection, String: "User Type", DefaultVal: "internal"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &ResUsers{}})
}
