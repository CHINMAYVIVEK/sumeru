package models

import "sumeru/core/base"

type StockPicking struct {
	base.BaseModel
	Name            string `db:"name"`
	PartnerName     string `db:"partner_name"`
	PickingTypeCode string `db:"picking_type_code"`
	State           string `db:"state"`
	ScheduledDate   string `db:"scheduled_date"`
	Origin          string `db:"origin"`
}

func (s *StockPicking) ModelName() string {
	return "stock.picking"
}

func (s *StockPicking) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, Required: true, String: "Reference"},
		{Name: "partner_name", Type: base.Char, String: "Contact"},
		{
			Name:       "picking_type_code",
			Type:       base.Selection,
			DefaultVal: "incoming",
			String:     "Operation Type",
			Index:      true,
		},
		{
			Name:       "state",
			Type:       base.Selection,
			DefaultVal: "draft",
			String:     "Status",
			Index:      true,
		},
		{Name: "scheduled_date", Type: base.Date, String: "Scheduled Date"},
		{Name: "origin", Type: base.Char, String: "Source Document"},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &StockPicking{}})
}
