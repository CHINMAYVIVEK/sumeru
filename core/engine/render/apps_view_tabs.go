package render

import (
	"net/url"
	"strings"
)

// AppsViewTabs builds Kanban / Grid / List links for the Apps dashboard (?layout=).
func AppsViewTabs(currentLayout, msg string) []ViewSwitchTab {
	cur := strings.ToLower(strings.TrimSpace(currentLayout))
	if cur == "" {
		cur = "grid"
	}
	msg = strings.TrimSpace(msg)
	order := []struct {
		layoutKey string
		label     string
		mode      string // template icon branch
	}{
		{"kanban", "Kanban", "apps_kanban"},
		{"grid", "Grid", "apps_grid"},
		{"list", "List", "apps_list"},
	}
	var out []ViewSwitchTab
	for _, o := range order {
		q := url.Values{}
		q.Set("layout", o.layoutKey)
		if msg != "" {
			q.Set("msg", msg)
		}
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   "/web/apps?" + q.Encode(),
			Mode:   o.mode,
			Active: cur == o.layoutKey,
		})
	}
	return out
}
