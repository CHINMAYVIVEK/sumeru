package render

import (
	"context"
	"html/template"
	"strings"

	"sumeru/core/engine/parser"
)

// UIHook allows addons to inject custom HTML into specific parts of the UI.
type UIHook func(ctx context.Context, vr *ViewRecordData, ro bool) template.HTML

var (
	// NotebookHooks: model -> page_title -> hook
	NotebookHooks = map[string]map[string]UIHook{}

	// ShellHooks are called for every page render to inject shell-level HTML.
	ShellHooks []UIHook
)

// RegisterShellHook adds a global hook to the shell rendering pipeline.
func RegisterShellHook(hook UIHook) {
	ShellHooks = append(ShellHooks, hook)
}

// RegisterNotebookHook registers a hook for a notebook page title on a model.
func RegisterNotebookHook(model, pageTitle string, hook UIHook) {
	if NotebookHooks[model] == nil {
		NotebookHooks[model] = map[string]UIHook{}
	}
	NotebookHooks[model][strings.ToLower(pageTitle)] = hook
}

// PageData is the top-level template payload for base.html.
type PageData struct {
	Title               string // legacy / diagnostics; prefer ViewBreadcrumb for UI
	ViewBreadcrumb      string // human label for breadcrumb (not the technical model id)
	AppName             string // product display name (browser tab suffix, header)
	ModuleName          string
	Content             template.HTML
	TopMenus            []parser.MenuItem
	SidebarMenus        []SidebarMenu
	ActiveModuleID      string
	ActiveMenuID        string
	ViewStylesheetURLs  []string
	AppsNavActive       bool
	SettingsNavActive   bool
	ExtraStylesheetURLs []string
	LogoURL             string
	// BrandLockupHref is the shell logo/name link target (default: home dashboard via EnrichShellPageData).
	BrandLockupHref string
	ShellCompany    string
	ShellUser       string
	UserInitial     string          // first letter for avatar
	ShellExtraHTML  template.HTML   // AI Assistant or other shell widgets
	ViewTabs        []ViewSwitchTab // workspace view switcher in breadcrumb bar; empty hides toolbar

	// BreadcrumbTrail: when non-empty, base.html renders linked crumbs; otherwise legacy ModuleName/ViewBreadcrumb.
	BreadcrumbItems []BreadcrumbItem

	// SuppressActivityDock forces the right activity dock off (e.g. Home dashboard) regardless of mail settings.
	SuppressActivityDock bool

	// ExtraBodyClasses is appended to the shell body class list (leading space recommended, e.g. " sum-body--settings-hub").
	ExtraBodyClasses string

	// Right activity panel: Log tab (audit); Messages tab HTML set in RenderView when chatter applies.
	ActivityEnabled         bool
	ActivityLogItems        []ActivityItem
	ActivityContextModel    string
	ActivityContextRecordID int64
	ActivityPanelChatter    bool
	ActivityChatterHTML     template.HTML
}

// ActivityItem is one line in the shell activity feed.
type ActivityItem struct {
	Meta string // author · relative time
	Body string
}

// SidebarMenu is a sidebar group with child menu links.
type SidebarMenu struct {
	ID       string
	Name     string
	Sequence int
	SubMenus []parser.MenuItem
}

// ViewRecordData carries rows loaded from the ORM for HTML rendering.
type ViewRecordData struct {
	ActionID int
	Record   map[string]interface{}
	ListRows []map[string]interface{}
	ViewTabs []ViewSwitchTab // optional; copied onto PageData for base layout

	// Workspace form chrome (/web): Edit / Save / Cancel and POST save target.
	ResModel       string // e.g. core.company
	RecordID       int    // 0 = create form
	FormEditing    bool   // true when URL contains edit=1
	FormBaseQuery  string // query string for /web without leading "?" and without edit= (action, menu_id, view_type, id)
	FormSaveAction string // POST URL; default "/web/record/save"
}
