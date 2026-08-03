package web

import (
	"net/http"
	"strings"

	"sumeru/core/orm"
)

// APIKeyUserID extracts a bearer / X-API-Key credential and resolves it to a user id.
func APIKeyUserID(r *http.Request) int {
	if r == nil {
		return 0
	}
	raw := strings.TrimSpace(r.Header.Get("X-API-Key"))
	if raw == "" {
		auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			raw = strings.TrimSpace(auth[7:])
		}
	}
	if raw == "" {
		return 0
	}
	return orm.UIDFromAPIKey(r.Context(), raw)
}

// AuthenticatedUserID returns session uid, or API key uid if no session.
func AuthenticatedUserID(r *http.Request) int {
	if uid := SessionUserID(r); uid > 0 {
		return uid
	}
	return APIKeyUserID(r)
}
