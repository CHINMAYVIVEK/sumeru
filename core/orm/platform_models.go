package orm

// SysAudit is an append-only audit trail (before/after JSON).
type SysAudit struct {
	ID         int    `orm:"id"`
	UserID     int    `orm:"user_id"`
	Action     string `orm:"action"` // create | write | unlink | login_fail | access_deny
	Model      string `orm:"model"`
	ResID      int64  `orm:"res_id"`
	BeforeJSON string `orm:"before_json"`
	AfterJSON  string `orm:"after_json"`
	Detail     string `orm:"detail"`
	CreateDate string `orm:"create_date"`
}

func (SysAudit) ModelName() string { return "sys.audit" }
func (SysAudit) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "user_id", Type: Many2One, Relation: "core.user", Index: true, String: "User"},
		{Name: "action", Type: Char, Required: true, Index: true, String: "Action"},
		{Name: "model", Type: Char, Index: true, String: "Model"},
		{Name: "res_id", Type: Integer, String: "Record"},
		{Name: "before_json", Type: Text, String: "Before"},
		{Name: "after_json", Type: Text, String: "After"},
		{Name: "detail", Type: Text, String: "Detail"},
		{Name: "create_date", Type: DateTime, Required: true, Index: true, String: "When"},
	}
}

// SysAttachment stores file metadata (content as base64/text or path).
type SysAttachment struct {
	ID         int    `orm:"id"`
	Name       string `orm:"name"`
	Model      string `orm:"model"`
	ResID      int64  `orm:"res_id"`
	Mimetype   string `orm:"mimetype"`
	FileSize   int64  `orm:"file_size"`
	Datas      string `orm:"datas"` // base64 or inline content for small files
	StoreFname string `orm:"store_fname"`
	CreateDate string `orm:"create_date"`
	CompanyID  int    `orm:"company_id"`
}

func (SysAttachment) ModelName() string { return "sys.attachment" }
func (SysAttachment) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "model", Type: Char, Index: true, String: "Model"},
		{Name: "res_id", Type: Integer, Index: true, String: "Record"},
		{Name: "mimetype", Type: Char, String: "MIME Type"},
		{Name: "file_size", Type: Integer, String: "Size"},
		{Name: "datas", Type: Text, String: "Data"},
		{Name: "store_fname", Type: Char, String: "Stored Filename"},
		{Name: "create_date", Type: DateTime, String: "Created"},
		{Name: "company_id", Type: Many2One, Relation: "core.company", String: "Company"},
	}
}

// SysSequence generates monotonic document numbers.
type SysSequence struct {
	ID       int    `orm:"id"`
	Name     string `orm:"name"`
	Code     string `orm:"code"`
	Prefix   string `orm:"prefix"`
	Suffix   string `orm:"suffix"`
	Padding  int    `orm:"padding"`
	NumberNext int  `orm:"number_next"`
	Active   bool   `orm:"active"`
}

func (SysSequence) ModelName() string { return "sys.sequence" }
func (SysSequence) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, String: "Name"},
		{Name: "code", Type: Char, Required: true, Unique: true, Index: true, String: "Code"},
		{Name: "prefix", Type: Char, String: "Prefix"},
		{Name: "suffix", Type: Char, String: "Suffix"},
		{Name: "padding", Type: Integer, DefaultVal: 5, String: "Padding"},
		{Name: "number_next", Type: Integer, DefaultVal: 1, String: "Next Number"},
		{Name: "active", Type: Boolean, DefaultVal: true, String: "Active"},
	}
}

// SysConfigParameter is a key/value system setting.
type SysConfigParameter struct {
	ID    int    `orm:"id"`
	Key   string `orm:"key"`
	Value string `orm:"value"`
}

func (SysConfigParameter) ModelName() string { return "sys.config_parameter" }
func (SysConfigParameter) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "key", Type: Char, Required: true, Unique: true, Index: true, String: "Key"},
		{Name: "value", Type: Text, String: "Value"},
	}
}

func init() {
	const kernel = "base"
	RegisterModelWithModule(SysAudit{}, kernel)
	RegisterModelWithModule(SysAttachment{}, kernel)
	RegisterModelWithModule(SysSequence{}, kernel)
	RegisterModelWithModule(SysConfigParameter{}, kernel)
}
