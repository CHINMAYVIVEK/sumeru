package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// RelSearchHandler GET /web/rel/search?model=core.partner&q=acme
// Optional: filter_field=country_id&filter_id=12 for cascade dropdowns.
func RelSearchHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	limit := 20
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			limit = n
		}
	}
	filterField := strings.TrimSpace(r.URL.Query().Get("filter_field"))
	var filterID int64
	if s := strings.TrimSpace(r.URL.Query().Get("filter_id")); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			filterID = n
		}
	}
	rows, err := orm.RelNameSearchFiltered(r.Context(), model, q, limit, filterField, filterID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, r.Context(), r.URL.Path, map[string]interface{}{"results": rows})
}
