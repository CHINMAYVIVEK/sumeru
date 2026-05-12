package models

import "sumeru/core/base"

type SaleOrder struct {
	base.BaseModel
	Name        string  `db:"name"`
	PartnerName string  `db:"partner_name"`
	DateOrder   string  `db:"date_order"`
	Email       string  `db:"email"`
	Phone       string  `db:"phone"`
	Amount      float64 `db:"amount"`
	Note        string  `db:"note"`
	State       string  `db:"state"`
	Priority    int     `db:"priority"`
	UserID      int64   `db:"user_id"`
	CompanyID   int64   `db:"company_id"`
}

func (s *SaleOrder) ModelName() string {
	return "sale.order"
}

func (s *SaleOrder) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{
			Name:     "name",
			Type:     base.Char,
			Required: true,
			String:   "Order Reference",
		},
		{
			Name:   "partner_name",
			Type:   base.Char,
			String: "Customer",
		},
		{
			Name:   "date_order",
			Type:   base.Date,
			String: "Order Date",
		},
		{
			Name:   "email",
			Type:   base.Char,
			String: "Email",
		},
		{
			Name:   "phone",
			Type:   base.Char,
			String: "Phone",
		},
		{
			Name:   "amount",
			Type:   base.Numeric,
			String: "Total",
		},
		{
			Name:   "note",
			Type:   base.Text,
			String: "Terms and conditions",
		},
		{
			Name:       "state",
			Type:       base.Selection,
			DefaultVal: "draft",
			String:     "Status",
			Index:      true,
		},
		{
			Name:   "priority",
			Type:   base.Integer,
			String: "Priority",
			Index:  true,
		},
		{
			Name:   "user_id",
			Type:   base.Integer,
			String: "Salesperson",
		},
		{
			Name:   "company_id",
			Type:   base.Integer,
			String: "Company",
		},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &SaleOrder{}})
}
