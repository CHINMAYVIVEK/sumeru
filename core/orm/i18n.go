package orm

import (
	"context"
	"database/sql"
	"strings"
)

// CoreLang is an installable language catalog entry.
type CoreLang struct {
	ID      int    `orm:"id"`
	Code    string `orm:"code"`
	Name    string `orm:"name"`
	Active  bool   `orm:"active"`
	ISOCode string `orm:"iso_code"`
}

func (CoreLang) ModelName() string { return "core.lang" }
func (CoreLang) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "code", Type: Char, Required: true, Unique: true, Index: true, String: "Code"},
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
		{Name: "iso_code", Type: Char, String: "ISO Code"},
	}
}

// SysTranslation stores a translated term for a language.
type SysTranslation struct {
	ID    int    `orm:"id"`
	Lang  string `orm:"lang"`
	Src   string `orm:"src"`
	Value string `orm:"value"`
	Module string `orm:"module"`
}

func (SysTranslation) ModelName() string { return "sys.translation" }
func (SysTranslation) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "lang", Type: Char, Required: true, Index: true, String: "Language"},
		{Name: "src", Type: Text, Required: true, String: "Source"},
		{Name: "value", Type: Text, String: "Translation"},
		{Name: "module", Type: Char, Index: true, String: "Module"},
	}
}

// SysFieldAccess is optional field-level security (read/write per group).
type SysFieldAccess struct {
	ID        int    `orm:"id"`
	Name      string `orm:"name"`
	Model     string `orm:"model"`
	FieldName string `orm:"field_name"`
	GroupID   int    `orm:"group_id"`
	PermRead  bool   `orm:"perm_read"`
	PermWrite bool   `orm:"perm_write"`
}

func (SysFieldAccess) ModelName() string { return "sys.field_access" }
func (SysFieldAccess) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Unique: true, String: "Name"},
		{Name: "model", Type: Char, Required: true, Index: true, String: "Model"},
		{Name: "field_name", Type: Char, Required: true, String: "Field"},
		{Name: "group_id", Type: Many2One, Relation: "core.group", String: "Group"},
		{Name: "perm_read", Type: Boolean, DefaultVal: true, String: "Read"},
		{Name: "perm_write", Type: Boolean, DefaultVal: true, String: "Write"},
	}
}

func init() {
	const kernel = "base"
	RegisterModelWithModule(CoreLang{}, kernel)
	RegisterModelWithModule(SysTranslation{}, kernel)
	RegisterModelWithModule(SysFieldAccess{}, kernel)
}

// Translate returns the translated value for src in lang, or src if missing.
func Translate(ctx context.Context, lang, src string) string {
	src = strings.TrimSpace(src)
	lang = strings.TrimSpace(lang)
	if src == "" || lang == "" || lang == "en_US" || DB == nil {
		return src
	}
	if _, ok := Registry["sys.translation"]; !ok {
		return src
	}
	var val sql.NullString
	err := DB.QueryRowContext(ContextWithBypass(ctx, true),
		`SELECT value FROM `+GetTableName("sys.translation")+
			` WHERE lang = $1 AND src = $2 LIMIT 1`, lang, src,
	).Scan(&val)
	if err != nil || !val.Valid || val.String == "" {
		return src
	}
	return val.String
}
