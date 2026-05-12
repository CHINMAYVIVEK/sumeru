package render

import (
	"bytes"
	"html/template"
	"path/filepath"
)

// RenderPage executes base.html with the given shell + content data.
func RenderPage(templatesDir string, data PageData) (string, error) {
	EnrichShellPageData(&data)
	tmpl, err := template.ParseFiles(
		filepath.Join(templatesDir, "base.html"),
		filepath.Join(templatesDir, "shell_partials.html"),
	)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	// Must execute the layout by name: the partials file only defines {{template "sumMenuIcon"}}.
	if err := tmpl.ExecuteTemplate(&buf, "base.html", data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
