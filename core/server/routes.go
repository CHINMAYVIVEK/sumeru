package server

import (
	"net/http"

	"sumeru/core/server/web"
)

func registerAppRoutes() {
	http.HandleFunc("/web/login", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			web.LoginGet(w, r)
		case http.MethodPost:
			web.LoginPost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/web/logout", web.LogoutGet)
	http.HandleFunc("/web", web.WebHandler)
	http.HandleFunc("/web/apps", web.AppsHandler)
	http.HandleFunc("/web/apps/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
	http.HandleFunc("/web/module/action", web.ModuleActionHandler)
	http.HandleFunc("/web/record/save", web.RecordSaveHandler)
	http.HandleFunc("/web/record/delete", web.RecordDeleteHandler)
	http.HandleFunc("/web/action/reset_password", web.ActionResetPassword)
	http.HandleFunc("/web/chatter/post", web.ChatterPostHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
}
