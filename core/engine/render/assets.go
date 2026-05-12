package render

// ExtraStylesheetURLs are linked after view stylesheets (e.g. /static/brand.css from deployment).
var ExtraStylesheetURLs []string

// SetExtraStylesheetURLs replaces the list of extra stylesheet URLs (absolute paths on site).
func SetExtraStylesheetURLs(urls []string) {
	ExtraStylesheetURLs = append([]string(nil), urls...)
}
