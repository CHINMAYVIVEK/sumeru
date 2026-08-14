// Package render builds workspace HTML from parsed views and ORM record data.
//
// Two rendering layers:
//   - Shell pages (page.go, shell_render.go) — html/template layout, top bar, sidebar
//   - View content (list_render.go, form_*.go, kanban_render.go) — strings.Builder widgets
//
// Escaping: escape all user/model values with template.HTMLEscapeString before WriteString.
// Shared field chrome and labels live in html_helpers.go.
//
// Extension: new field widgets register in form_widgets.go (or form_widget_*.go);
// model-specific overrides (e.g. core.user selects) stay in dedicated files until hooks exist.
package render
