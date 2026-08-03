package models

import "sumeru/core/sdk"

// SysTranslation stores a translated term for a language.
type SysTranslation struct {
	sdk.BaseModel
	Lang   string `db:"lang"`
	Src    string `db:"src"`
	Value  string `db:"value"`
	Module string `db:"module"`
}

func (SysTranslation) ModelName() string { return "sys.translation" }

func (SysTranslation) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "lang", Type: sdk.Char, Required: true, Index: true, String: "Language"},
		{Name: "src", Type: sdk.Text, Required: true, String: "Source"},
		{Name: "value", Type: sdk.Text, String: "Translation"},
		{Name: "module", Type: sdk.Char, Index: true, String: "Module"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &SysTranslation{}, Module: "base"})
}
