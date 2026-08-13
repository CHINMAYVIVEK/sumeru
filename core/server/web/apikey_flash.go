package web

import (
	"net/http"

	"sumeru/core/server/config"
)

const apiKeyFlashCookie = "sumeru_api_key_flash"

// SetAPIKeyFlash stores a one-time raw API key in an HttpOnly cookie (never in redirect URLs).
func SetAPIKeyFlash(w http.ResponseWriter, raw string) {
	http.SetCookie(w, &http.Cookie{
		Name:     apiKeyFlashCookie,
		Value:    raw,
		Path:     "/",
		MaxAge:   120,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	})
}

// ConsumeAPIKeyFlash reads and clears the one-time API key flash cookie.
func ConsumeAPIKeyFlash(r *http.Request, w http.ResponseWriter) string {
	c, err := r.Cookie(apiKeyFlashCookie)
	http.SetCookie(w, &http.Cookie{
		Name:     apiKeyFlashCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   !config.AppConfig.DevMode,
	})
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}
