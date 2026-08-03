package sdk

import "sumeru/core/orm"

// Model and field types mirror the ORM; they are type aliases so struct tags and
// method signatures stay identical while callers import only sdk.

type (
	Model           = orm.Model
	FieldDefinition = orm.FieldDefinition
	FieldType       = orm.FieldType
	BaseModel       = orm.BaseModel
)

const (
	Char      = orm.Char
	Text      = orm.Text
	Integer   = orm.Integer
	Float     = orm.Float
	Numeric   = orm.Numeric
	Boolean   = orm.Boolean
	Date      = orm.Date
	DateTime  = orm.DateTime
	Selection = orm.Selection
	Many2One  = orm.Many2One
	Many2Many = orm.Many2Many
	One2Many  = orm.One2Many
	Json      = orm.Json
)
