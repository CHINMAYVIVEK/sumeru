package models

import "sumeru/core/base"

type ProductProduct struct {
	base.BaseModel
	Name         string  `db:"name"`
	DefaultCode  string  `db:"default_code"`
	ListPrice    float64 `db:"list_price"`
	QtyAvailable float64 `db:"qty_available"`
	Active       bool    `db:"active"`
}

func (p *ProductProduct) ModelName() string {
	return "product.product"
}

func (p *ProductProduct) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Product Name"},
		{Name: "default_code", Type: base.Char, String: "Internal Reference", Index: true},
		{Name: "list_price", Type: base.Numeric, String: "Sales Price"},
		{Name: "qty_available", Type: base.Numeric, String: "Quantity On Hand"},
		{Name: "active", Type: base.Boolean, DefaultVal: true, String: "Active"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &ProductProduct{}})
}
