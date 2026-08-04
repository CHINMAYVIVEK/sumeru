package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// RelSearchHandler GET /web/rel/search?model=core.partner&q=acme
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
	rows, err := orm.RelNameSearch(r.Context(), model, q, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": rows})
}
