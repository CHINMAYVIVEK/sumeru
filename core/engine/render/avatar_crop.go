package render

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
)

// ImageCrop holds normalized pan/zoom for circular avatar framing (form + shell).
type ImageCrop struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Zoom float64 `json:"zoom"`
}

// ParseImageCrop decodes image_crop JSON; returns defaults when empty or invalid.
func ParseImageCrop(raw string) (ImageCrop, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ImageCrop{X: 50, Y: 50, Zoom: 1}, false
	}
	var c ImageCrop
	if err := json.Unmarshal([]byte(raw), &c); err != nil {
		return ImageCrop{X: 50, Y: 50, Zoom: 1}, false
	}
	return NormalizeImageCrop(c), true
}

// NormalizeImageCrop clamps crop values to supported ranges.
func NormalizeImageCrop(c ImageCrop) ImageCrop {
	if c.X < 0 {
		c.X = 0
	} else if c.X > 100 {
		c.X = 100
	}
	if c.Y < 0 {
		c.Y = 0
	} else if c.Y > 100 {
		c.Y = 100
	}
	if c.Zoom < 1 {
		c.Zoom = 1
	} else if c.Zoom > 4 {
		c.Zoom = 4
	}
	return c
}

// AvatarCropStyle returns inline CSS for object-position + zoom inside a circular frame.
func AvatarCropStyle(c ImageCrop, active bool) template.HTMLAttr {
	if !active {
		return template.HTMLAttr("")
	}
	c = NormalizeImageCrop(c)
	style := fmt.Sprintf(
		`object-position:%.2f%% %.2f%%;transform:scale(%.3f);transform-origin:%.2f%% %.2f%%`,
		c.X, c.Y, c.Zoom, c.X, c.Y,
	)
	return template.HTMLAttr(` style="` + template.HTMLEscapeString(style) + `"`)
}
