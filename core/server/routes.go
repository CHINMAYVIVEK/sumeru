package server

import (
	"net/http"

	"sumeru/core/server/web"
)

func registerAppRoutes() {
	http.HandleFunc("/web", web.WebHandler)
	http.HandleFunc("/web/apps", web.AppsHandler)
	http.HandleFunc("/web/apps/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
	http.HandleFunc("/web/module/action", web.ModuleActionHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/web/apps", http.StatusFound)
	})
}
