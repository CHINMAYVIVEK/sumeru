package models

import "sumeru/core/base"

type CoreCompany struct {
	base.BaseModel
	Name                     string `db:"name"`
	Street                   string `db:"street"`
	Street2                  string `db:"street2"`
	City                     string `db:"city"`
	Zip                      string `db:"zip"`
	State                    string `db:"state"`
	Country                  string `db:"country"`
	Email                    string `db:"email"`
	Phone                    string `db:"phone"`
	Mobile                   string `db:"mobile"`
	Website                  string `db:"website"`
	Vat                      string `db:"vat"`
	CompanyRegistry          string `db:"company_registry"`
	InternalNotes            string `db:"internal_notes"`
	MailChatterEnabled       bool   `db:"mail_chatter_enabled"`
	MailActivityPanelEnabled bool   `db:"mail_activity_panel_enabled"`
}

func (CoreCompany) ModelName() string { return "core.company" }

func (CoreCompany) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, Unique: true, String: "Company Name", Index: true},
		{Name: "street", Type: base.Char, String: "Street"},
		{Name: "street2", Type: base.Char, String: "Street 2"},
		{Name: "city", Type: base.Char, String: "City"},
		{Name: "zip", Type: base.Char, String: "Zip"},
		{Name: "state", Type: base.Char, String: "State / Province"},
		{Name: "country", Type: base.Char, String: "Country"},
		{Name: "email", Type: base.Char, String: "Email"},
		{Name: "phone", Type: base.Char, String: "Phone"},
		{Name: "mobile", Type: base.Char, String: "Mobile"},
		{Name: "website", Type: base.Char, String: "Website"},
		{Name: "vat", Type: base.Char, String: "Tax ID", Index: true},
		{Name: "company_registry", Type: base.Char, String: "Company ID"},
		{Name: "internal_notes", Type: base.Text, String: "Internal Notes"},
		{Name: "mail_chatter_enabled", Type: base.Boolean, String: "Chatter", DefaultVal: true},
		{Name: "mail_activity_panel_enabled", Type: base.Boolean, String: "Activity panel", DefaultVal: true},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &CoreCompany{}, Module: "base"})
}
