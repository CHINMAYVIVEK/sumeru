// Package web implements HTTP handlers for the signed-in UI (/web/*), setup, and JSON APIs.
//
// Handler files (by concern):
//   - home_dashboard.go — app hub at /web/home
//   - workspace.go, workspace_load.go, workspace_resolve.go — main /web workspace
//   - apps_list.go, apps_detail.go, apps_state.go, apps_module_actions.go — Apps admin
//   - settings_hub.go, settings_app_logs.go — Settings overview and logs
//   - auth.go, auth_session.go, auth_helpers.go — login, session, access guards
//   - module_tile_helpers.go, module_parse_helpers.go, view_layout_helpers.go — shared helpers
package web

import (
	"net/http"

	"sumeru/core/server/router"
)

func init() {
	// Bridge auth helpers into the router registry (used by AuthSession / AuthAPIKey).
	router.RequireSession = func(w http.ResponseWriter, r *http.Request) bool {
		return requireLogin(w, r)
	}
	router.ResolveUID = func(r *http.Request) int {
		if uid := SessionUserID(r); uid > 0 {
			return uid
		}
		return APIKeyUserID(r)
	}
}

// RegisterAppRoutes registers HTTP handlers for /web and related paths after DB init.
//
// Registry routes (router.Register → router.Apply; addon-overridable):
//   - Auth: GET/POST /web/login, GET /web/logout
//   - Session: GET /web/home (app hub), GET /web (workspace views), GET /web/apps
//   - Session: GET /web/settings, GET /web/settings/app-logs, GET /web/rel/search
//   - Session POST: /web/company/switch, /web/module/action, /web/record/save|delete
//   - Session POST: /web/action/{reset_password,create_api_key,object}, /web/chatter/post, /web/import/csv
//   - API: GET /api/health (none), POST /api/rpc (handler enforces session or API key)
//   - Session + base.group_system: GET /metrics
//
// Alias redirects (direct mux; legacy/trailing-slash normalization):
//   - /web/apps/ → /web/apps
//   - /web/settings/… → /web/settings (prefix; /web/settings/app-logs stays exact via registry)
//   - / → /web/home
func RegisterAppRoutes(mux *http.ServeMux) {
	reg := mux
	if reg == nil {
		reg = http.DefaultServeMux
	}

	router.Register(http.MethodGet, "/web/login", router.AuthNone, LoginGet)
	router.Register(http.MethodPost, "/web/login", router.AuthNone, LoginPost)
	router.Register(http.MethodPost, "/web/company/switch", router.AuthSession, SwitchCompanyPost)
	router.Register(http.MethodGet, "/web/logout", router.AuthNone, LogoutGet)
	router.Register(http.MethodGet, "/web/home", router.AuthSession, HomeDashboardHandler)
	router.Register(http.MethodPost, "/web/user/pinned-apps", router.AuthSession, PinnedAppsSaveHandler)
	router.Register(http.MethodGet, "/web", router.AuthSession, WebHandler)
	router.Register(http.MethodGet, "/web/apps", router.AuthSession, AppsHandler)
	router.Register(http.MethodPost, "/web/module/action", router.AuthSession, ModuleActionHandler)
	router.Register(http.MethodPost, "/web/record/save", router.AuthSession, RecordSaveHandler)
	router.Register(http.MethodPost, "/web/record/delete", router.AuthSession, RecordDeleteHandler)
	router.Register(http.MethodGet, "/web/rel/search", router.AuthSession, RelSearchHandler)
	router.Register(http.MethodPost, "/web/action/reset_password", router.AuthSession, ActionResetPassword)
	router.Register(http.MethodPost, "/web/action/create_api_key", router.AuthSession, ActionCreateAPIKey)
	router.Register(http.MethodPost, "/web/action/object", router.AuthSession, ActionObjectHandler)
	router.Register(http.MethodPost, "/web/kanban/move", router.AuthSession, KanbanMoveHandler)
	router.Register(http.MethodPost, "/web/chatter/post", router.AuthSession, ChatterPostHandler)
	router.Register(http.MethodGet, "/web/settings", router.AuthSession, SettingsHubHandler)
	router.Register(http.MethodGet, "/web/settings/app-logs", router.AuthSession, AppLogsHandler)
	router.Register(http.MethodGet, "/api/health", router.AuthNone, APIHealthHandler)
	router.Register(http.MethodPost, "/api/rpc", router.AuthNone, RPCJSONHandler)
	router.Register(http.MethodGet, "/metrics", router.AuthSession, MetricsHandler)
	router.Register(http.MethodPost, "/web/import/csv", router.AuthSession, ImportCSVHandler)

	router.Apply(reg)
	registerAppAliases(reg)
}

func redirectTo(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target, http.StatusFound)
	}
}

func rootRedirectApp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/web/home", http.StatusFound)
}

func registerAppAliases(mux *http.ServeMux) {
	mux.HandleFunc("/web/apps/", redirectTo("/web/apps"))
	mux.HandleFunc("/web/settings/", redirectTo("/web/settings"))
	mux.HandleFunc("/", rootRedirectApp)
}

// RegisterSetupRoutes registers first-run setup handlers (DB not ready).
func RegisterSetupRoutes(mux *http.ServeMux) {
	reg := mux
	if reg == nil {
		reg = http.DefaultServeMux
	}
	reg.HandleFunc("/setup", SetupPageHandler)
	reg.HandleFunc("/setup/init", SetupInitHandler)

	reg.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/setup", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
}
