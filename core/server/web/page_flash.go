package web

import (
	"net/http"
)

// PageFlash is a one-time user-visible banner after redirect.
type PageFlash struct {
	Kind  string // success, info, warning
	Title string
	Body  string
}

// ConsumePageFlashes reads and clears one-time flash data (cookies).
func ConsumePageFlashes(r *http.Request, w http.ResponseWriter) []PageFlash {
	var out []PageFlash
	if key := ConsumeAPIKeyFlash(r, w); key != "" {
		out = append(out, PageFlash{
			Kind:  "success",
			Title: "API key created",
			Body:  "Copy this key now — it will not be shown again:\n" + key,
		})
	}
	return out
}
