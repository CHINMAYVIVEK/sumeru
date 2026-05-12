package base

import "sumeru/core/engine/render"

// SetExtraStylesheetURLsInput registers extra CSS URLs for the shell layout.
type SetExtraStylesheetURLsInput struct {
	URLs []string
}

// SetExtraStylesheetURLs forwards to the render layer (e.g. brand_css route).
func SetExtraStylesheetURLs(in SetExtraStylesheetURLsInput) {
	render.SetExtraStylesheetURLs(in.URLs)
}
