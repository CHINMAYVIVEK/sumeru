package assets

// DefaultStylesheetURLs is the ordered list of core CSS files for the web UI.
// sumeru-ai.css is omitted; append via AIStylesheetURL when the sumeru_ai module is installed.
func DefaultStylesheetURLs() []string {
	return []string{
		"/static/css/sumeru-theme.css",
		"/static/css/sumeru-base.css",
		"/static/css/sumeru-shell.css",
		"/static/css/sumeru-messages.css",
		"/static/css/sumeru-views.css",
		"/static/css/sumeru-compat.css",
		"/static/css/sumeru-login.css",
		"/static/css/sumeru-pages.css",
	}
}

// AIStylesheetURL is the optional AI assistant stylesheet path.
func AIStylesheetURL() string {
	return "/static/css/sumeru-ai.css"
}
