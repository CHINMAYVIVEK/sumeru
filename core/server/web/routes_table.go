package web

import (
	"net/http"

	"sumeru/core/server/controller"
)

func init() {
	// Bridge auth helpers into the controller registry (used by AuthSession / AuthAPIKey).
	controller.RequireSession = func(w http.ResponseWriter, r *http.Request) bool {
		return requireLogin(w, r)
	}
	controller.ResolveUID = func(r *http.Request) int {
		if uid := SessionUserID(r); uid > 0 {
			return uid
		}
		return APIKeyUserID(r)
	}
}

// RegisterAppRoutes registers HTTP handlers for /web and related paths after DB init.
// Built-in routes go through the controller registry so addons can override/extend.
func RegisterAppRoutes(mux *http.ServeMux) {
	reg := mux
	if reg == nil {
		reg = http.DefaultServeMux
	}

	controller.Register(http.MethodGet, "/web/login", controller.AuthNone, LoginGet)
	controller.Register(http.MethodPost, "/web/login", controller.AuthNone, LoginPost)
	controller.Register(http.MethodPost, "/web/company/switch", controller.AuthSession, SwitchCompanyPost)
	controller.Register(http.MethodGet, "/web/logout", controller.AuthNone, LogoutGet)
	controller.Register(http.MethodGet, "/web/home", controller.AuthSession, HomeDashboardHandler)
	controller.Register(http.MethodGet, "/web", controller.AuthSession, WebHandler)
	controller.Register(http.MethodGet, "/web/apps", controller.AuthSession, AppsHandler)
	controller.Register(http.MethodPost, "/web/module/action", controller.AuthSession, ModuleActionHandler)
	controller.Register(http.MethodPost, "/web/record/save", controller.AuthSession, RecordSaveHandler)
	controller.Register(http.MethodPost, "/web/record/delete", controller.AuthSession, RecordDeleteHandler)
	controller.Register(http.MethodPost, "/web/action/reset_password", controller.AuthSession, ActionResetPassword)
	controller.Register(http.MethodPost, "/web/chatter/post", controller.AuthSession, ChatterPostHandler)
	controller.Register(http.MethodGet, "/web/settings", controller.AuthSession, SettingsHubHandler)
	controller.Register(http.MethodGet, "/web/settings/app-logs", controller.AuthSession, AppLogsHandler)
	controller.Register(http.MethodGet, "/api/health", controller.AuthNone, APIHealthHandler)
	controller.Register(http.MethodPost, "/api/rpc", controller.AuthAPIKey, RPCJSONHandler)
	controller.Register(http.MethodPost, "/web/import/csv", controller.AuthSession, ImportCSVHandler)

	controller.Apply(reg)

	reg.HandleFunc("/web/apps/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
	reg.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/web/home", http.StatusFound)
	})
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
