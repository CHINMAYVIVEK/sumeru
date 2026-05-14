package orm

type FieldType string

const (
	Char      FieldType = "char"
	Text      FieldType = "text"
	Integer   FieldType = "integer"
	Float     FieldType = "float"
	Numeric   FieldType = "numeric" // For exact decimal precision (money)
	Boolean   FieldType = "boolean"
	Date      FieldType = "date"
	DateTime  FieldType = "datetime"
	Selection FieldType = "selection"
	Many2One  FieldType = "many2one"
	Many2many FieldType = "many2many"
	One2many  FieldType = "one2many"
	Json      FieldType = "json"
)

type FieldDefinition struct {
	Name       string
	Type       FieldType
	Required   bool
	Relation   string // For Many2One, Many2many, One2many
	RelationTable string // For Many2many
	Column1    string // For Many2many (this model's FK in rel table)
	Column2    string // For Many2many (target model's FK in rel table)
	String     string // Label
	DefaultVal interface{}
	Unique     bool
	Index      bool // Generate database index
}

type Model interface {
	ModelName() string
	Fields() []FieldDefinition
}

type BaseModel struct {
	ID int `db:"id"`
}
