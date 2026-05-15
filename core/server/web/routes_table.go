package web

import "net/http"

// RegisterAppRoutes registers HTTP handlers for /web and related paths after DB init.
// If mux is nil, handlers are registered on [http.DefaultServeMux].
func RegisterAppRoutes(mux *http.ServeMux) {
	reg := mux
	if reg == nil {
		reg = http.DefaultServeMux
	}
	reg.HandleFunc("/web/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			LoginGet(w, r)
		case http.MethodPost:
			LoginPost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	reg.HandleFunc("/web/logout", LogoutGet)
	reg.HandleFunc("/web/home", HomeDashboardHandler)
	reg.HandleFunc("/web", WebHandler)
	reg.HandleFunc("/web/apps", AppsHandler)
	reg.HandleFunc("/web/apps/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
	reg.HandleFunc("/web/module/action", ModuleActionHandler)
	reg.HandleFunc("/web/record/save", RecordSaveHandler)
	reg.HandleFunc("/web/record/delete", RecordDeleteHandler)
	reg.HandleFunc("/web/action/reset_password", ActionResetPassword)
	reg.HandleFunc("/web/chatter/post", ChatterPostHandler)

	reg.HandleFunc("/web/settings", SettingsHubHandler)
	reg.HandleFunc("/web/settings/app-logs", AppLogsHandler)

	reg.HandleFunc("/api/health", APIHealthHandler)
	reg.HandleFunc("/api/rpc", RPCJSONHandler)

	reg.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/web/home", http.StatusFound)
	})
}

// RegisterSetupRoutes registers first-run setup handlers (DB not ready).
// If mux is nil, handlers are registered on [http.DefaultServeMux].
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
