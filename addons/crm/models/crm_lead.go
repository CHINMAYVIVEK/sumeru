package models

import "sumeru/core/base"

type CrmLead struct {
	base.BaseModel
	Name            string  `db:"name"`
	PartnerName     string  `db:"partner_name"`
	Email           string  `db:"email"`
	Phone           string  `db:"phone"`
	Stage           string  `db:"stage"`
	ExpectedRevenue float64 `db:"expected_revenue"`
	Probability     float64 `db:"probability"`
	UserID          int64   `db:"user_id"`
	Description     string  `db:"description"`
}

func (c *CrmLead) ModelName() string {
	return "crm.lead"
}

func (c *CrmLead) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Opportunity"},
		{Name: "partner_name", Type: base.Char, String: "Contact"},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{
			Name:       "stage",
			Type:       base.Selection,
			DefaultVal: "new",
			String:     "Stage",
			Index:      true,
		},
		{Name: "expected_revenue", Type: base.Numeric, String: "Expected Revenue"},
		{Name: "probability", Type: base.Numeric, String: "Probability"},
		{Name: "user_id", Type: base.Integer, String: "Salesperson"},
		{Name: "description", Type: base.Text, String: "Internal Notes"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CrmLead{}})
}
