package assets

// DefaultStylesheetURLs is the ordered list of core CSS files for the web UI.
// sumeru-theme.css must load first (design tokens). Structural sheets use var(--sum-*).
func DefaultStylesheetURLs() []string {
	return []string{
		"/static/css/sumeru-theme.css",
		"/static/css/sumeru-base.css",
		"/static/css/sumeru-shell.css",
		"/static/css/sumeru-messages.css",
		"/static/css/sumeru-views.css",
		"/static/css/sumeru-compat.css",
		"/static/css/sumeru-ai.css",
		"/static/css/sumeru-login.css",
		"/static/css/sumeru-pages.css",
	}
}
