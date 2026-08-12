package models

import "sumeru/core/sdk"

type CoreUserAPIKey struct {
	sdk.BaseModel
	UserID     int    `db:"user_id"`
	Name       string `db:"name"`
	KeyPrefix  string `db:"key_prefix"`
	KeyHash    string `db:"key_hash"`
	Active     bool   `db:"active"`
	CreateDate string `db:"create_date"`
}

func (CoreUserAPIKey) ModelName() string { return "core.user.apikey" }

func (CoreUserAPIKey) Fields() []sdk.FieldDefinition {
	return []sdk.FieldDefinition{
		{Name: "user_id", Type: sdk.Many2One, Relation: "core.user", Required: true, Index: true, String: "User"},
		{Name: "name", Type: sdk.Char, Required: true, String: "Name"},
		{Name: "key_prefix", Type: sdk.Char, Required: true, String: "Prefix"},
		{Name: "key_hash", Type: sdk.Char, Required: true, String: "Hash"},
		{Name: "active", Type: sdk.Boolean, DefaultVal: true, String: "Active"},
		{Name: "create_date", Type: sdk.DateTime, Required: true, String: "Created"},
	}
}

func init() {
	sdk.RegisterModel(sdk.RegisterModelInput{Model: &CoreUserAPIKey{}, Module: "base"})
}
