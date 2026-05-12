package config

const (
	defaultDbSSLMode = "disable"

	iniSectionPrefix = "["
	iniCommentSemi   = ";"
	iniCommentHash   = "#"
	iniSeparator     = "="
	iniKeyValueParts = 2

	keyDbHost             = "db_host"
	keyDbPort             = "db_port"
	keyDbUser             = "db_user"
	keyDbPassword         = "db_password"
	keyDbName             = "db_name"
	keyDbSSLMode          = "db_sslmode"
	keyHTTPPort           = "http_port"
	keyAddonsPath         = "addons_path"
	keySumeruHome         = "sumeru_home"
	keyAssetsPath         = "assets_path"
	keyTemplatesPath      = "templates_path"
	keyBrandCSS           = "brand_css"
	keyLogoPath           = "logo_path"
	keyCompanyDisplayName = "company_display_name"
	keyUserDisplayName    = "user_display_name"
	keyLogFile            = "log_file"
)

const (
	relPathDefaultAssets    = "core/engine/assets"
	relPathDefaultTemplates = "core/engine/templates"
)

const (
	fileGoMod = "go.mod"

	segCore      = "core"
	segBase      = "base"
	segEngine    = "engine"
	segAssets    = "assets"
	segTemplates = "templates"
)

const addonsPathDelimiter = ","

const (
	errFmtDbHostRequired     = "db_host is required in %s"
	errFmtDbPortRequired     = "db_port is required in %s"
	errFmtDbUserRequired     = "db_user is required in %s"
	errFmtDbPasswordRequired = "db_password is required in %s"
	errFmtDbNameRequired     = "db_name is required in %s"
	errFmtHTTPPortRequired   = "http_port is required in %s"
	errFmtAddonsPathRequired = "addons_path is required in %s"
)
