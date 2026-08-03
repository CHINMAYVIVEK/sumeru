package models

import "sumeru/core/sdk"

type CoreCompany struct {
	sdk.BaseModel
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

func (CoreCompany) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "name", Type: sdk.Char, Required: true, Unique: true, String: "Company Name", Index: true},
		{Name: "street", Type: sdk.Char, String: "Street"},
		{Name: "street2", Type: sdk.Char, String: "Street 2"},
		{Name: "city", Type: sdk.Char, String: "City"},
		{Name: "zip", Type: sdk.Char, String: "Zip"},
		{Name: "state", Type: sdk.Char, String: "State / Province"},
		{Name: "country", Type: sdk.Char, String: "Country"},
		{Name: "email", Type: sdk.Char, String: "Email"},
		{Name: "phone", Type: sdk.Char, String: "Phone"},
		{Name: "mobile", Type: sdk.Char, String: "Mobile"},
		{Name: "website", Type: sdk.Char, String: "Website"},
		{Name: "vat", Type: sdk.Char, String: "Tax ID", Index: true},
		{Name: "company_registry", Type: sdk.Char, String: "Company ID"},
		{Name: "internal_notes", Type: sdk.Text, String: "Internal Notes"},
		{Name: "mail_chatter_enabled", Type: sdk.Boolean, String: "Chatter", DefaultVal: true},
		{Name: "mail_activity_panel_enabled", Type: sdk.Boolean, String: "Activity panel", DefaultVal: true},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreCompany{}, Module: "base"})
}
