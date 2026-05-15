package server

import (
	"net/http"

	"sumeru/core/server/web"
)

func registerAppRoutes() {
	http.HandleFunc("/web/login", func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodGet:
			web.LoginGet(responseWriter, request)
		case http.MethodPost:
			web.LoginPost(responseWriter, request)
		default:
			http.Error(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/web/logout", web.LogoutGet)
	http.HandleFunc("/web/home", web.HomeDashboardHandler)
	http.HandleFunc("/web", web.WebHandler)
	http.HandleFunc("/web/apps", web.AppsHandler)
	http.HandleFunc("/web/apps/", func(responseWriter http.ResponseWriter, request *http.Request) {
		http.Redirect(responseWriter, request, "/web/apps", http.StatusFound)
	})
	http.HandleFunc("/web/module/action", web.ModuleActionHandler)
	http.HandleFunc("/web/record/save", web.RecordSaveHandler)
	http.HandleFunc("/web/record/delete", web.RecordDeleteHandler)
	http.HandleFunc("/web/action/reset_password", web.ActionResetPassword)
	http.HandleFunc("/web/chatter/post", web.ChatterPostHandler)

	// Settings routes
	http.HandleFunc("/web/settings", web.SettingsHubHandler)
	http.HandleFunc("/web/settings/app-logs", web.AppLogsHandler)

	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/" {
			http.NotFound(responseWriter, request)
			return
		}
		http.Redirect(responseWriter, request, "/web/home", http.StatusFound)
	})
}

func registerSetupRoutes() {
	http.HandleFunc("/setup", web.SetupPageHandler)
	http.HandleFunc("/setup/init", web.SetupInitHandler)

	// Redirect root to setup, but only if it's exactly '/'
	http.HandleFunc("/", func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/" {
			http.Redirect(responseWriter, request, "/setup", http.StatusFound)
			return
		}
		// Fallback for other paths (not handled by static or other setup routes)
		http.NotFound(responseWriter, request)
	})
}
