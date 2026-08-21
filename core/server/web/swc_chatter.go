package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"sumeru/addons/mail"
)

const swcChatterRoute = "/web/swc/chatter"

func registerSwcChatterRoute() {
	registerSession(http.MethodGet, swcChatterRoute, SwcChatterHandler)
}

type swcChatterMessage struct {
	Body       string `json:"body"`
	Author     string `json:"author"`
	CreateDate string `json:"createDate"`
	Subtype    string `json:"subtype"`
}

type swcChatterPayload struct {
	Model    string              `json:"model"`
	RecordID int64               `json:"recordId"`
	Messages []swcChatterMessage `json:"messages"`
	Enabled  bool                `json:"enabled"`
}

// SwcChatterHandler GET /web/swc/chatter?model=&id=
func SwcChatterHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	if !mail.CompanyChatterEnabled(r.Context()) {
		writeJSONResponse(w, swcChatterPayload{Enabled: false})
		return
	}
	model := strings.TrimSpace(r.URL.Query().Get("model"))
	idRaw := strings.TrimSpace(r.URL.Query().Get("id"))
	id, _ := strconv.ParseInt(idRaw, 10, 64)
	if model == "" || id <= 0 {
		http.Error(w, "model and id required", http.StatusBadRequest)
		return
	}
	rows, err := mail.ListCommentsForRecord(r.Context(), model, id, 120)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	out := swcChatterPayload{
		Model:    model,
		RecordID: id,
		Enabled:  true,
	}
	for _, row := range rows {
		out.Messages = append(out.Messages, swcChatterMessage{
			Body:       row.Body,
			Author:     row.Author,
			CreateDate: row.CreateDate.UTC().Format("2006-01-02 15:04:05"),
			Subtype:    row.Subtype,
		})
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(out)
}
