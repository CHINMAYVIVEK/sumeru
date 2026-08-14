package web

// Shared routes used across web handlers.
const (
	loginRoute          = "/web/login"
	homeRoute           = "/web/home"
	companySwitchRoute  = "/web/company/switch"
	kanbanMoveRoute     = "/web/kanban/move"
	moduleActionRoute   = "/web/module/action"
	resetPasswordRoute  = "/web/action/reset_password"
)

// ORM model names referenced by web handlers.
const (
	sysMenuModel     = "sys.menu"
	coreCompanyModel = "core.company"
	coreUserModel    = "core.user"
)

// ACL group XML ids.
const groupSystemXML = "base.group_system"

// Common HTTP response messages.
const (
	forbiddenMessage   = "Forbidden"
	invalidCSRFMessage = "Invalid CSRF token"
)

// Home dashboard page identifiers.
const (
	homeInnerTemplate = "home_dashboard_inner.html"
	homePageTitle     = "Home"
	homeStylesheetURL = "/static/css/sumeru-home.css"
	homeEmptyMessage  = "No installed applications. Install apps from Apps."
	baseModuleName    = "base"
)

// Apps module action form fields (POST apps_*).
const (
	appsLayoutField  = "apps_layout"
	appsFilterField  = "apps_filter"
	appsScopeField   = "apps_scope"
	appsSearchField  = "apps_q"
)

// Apps module lifecycle action names (POST do=).
const (
	moduleActionInstall     = "install"
	moduleActionUninstall   = "uninstall"
	moduleActionDeactivate  = "deactivate"
	moduleActionActivate    = "activate"
	moduleActionSaveModule  = "save_module"
)

// Kanban move field names.
const (
	stageIDField              = "stage_id"
	dateLastStageUpdateField  = "date_last_stage_update"
)

// Company switch form field.
const companyIDFormField = "company_id"
