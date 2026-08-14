package web

import (
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// RelSearchHandler serves GET /web/rel/search for many2one typeahead and cascade filters.
func RelSearchHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}

	request := parseRelSearchRequest(r)
	rows, err := orm.RelNameSearchFiltered(
		r.Context(),
		request.ModelName,
		request.Query,
		request.Limit,
		request.FilterField,
		request.FilterID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, r.Context(), relSearchRoute, map[string]interface{}{"results": rows})
}

type relSearchRequest struct {
	ModelName   string
	Query       string
	Limit       int
	FilterField string
	FilterID    int64
}

func parseRelSearchRequest(r *http.Request) relSearchRequest {
	query := r.URL.Query()
	return relSearchRequest{
		ModelName:   strings.TrimSpace(query.Get(recordModelField)),
		Query:       strings.TrimSpace(query.Get(relSearchQueryParam)),
		Limit:       queryIntOrDefault(query.Get(relSearchLimitParam), defaultRelSearchLimit),
		FilterField: strings.TrimSpace(query.Get(relSearchFilterFieldParam)),
		FilterID:    queryInt64OrZero(query.Get(relSearchFilterIDParam)),
	}
}

func queryIntOrDefault(raw string, defaultValue int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}
	return parsed
}

func queryInt64OrZero(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
