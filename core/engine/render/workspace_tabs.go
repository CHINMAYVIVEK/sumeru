package render

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"sumeru/core/orm"
)

// WorkspaceViewTabs builds URLs for each view mode that has a default sys.view for resModel.
func WorkspaceViewTabs(ctx context.Context, resModel string, actionID int, menuID, selectedMode, recordID string) []ViewSwitchTab {
	order := []struct {
		mode  string
		label string
	}{
		{"kanban", "Kanban"},
		{"list", "List"},
		{"form", "Form"},
	}
	sel := strings.ToLower(strings.TrimSpace(selectedMode))
	menuID = strings.TrimSpace(menuID)
	recID := strings.TrimSpace(recordID)

	var out []ViewSwitchTab
	for _, o := range order {
		if _, err := orm.FindUIDefaultView(ctx, resModel, o.mode); err != nil {
			continue
		}
		q := url.Values{}
		q.Set("action", fmt.Sprintf("%d", actionID))
		if menuID != "" {
			q.Set("menu_id", menuID)
		}
		q.Set("view_type", o.mode)
		if o.mode == "form" && recID != "" {
			q.Set("id", recID)
		}
		href := "/web?" + q.Encode()
		out = append(out, ViewSwitchTab{
			Label:  o.label,
			Href:   href,
			Mode:   o.mode,
			Active: sel == o.mode,
		})
	}
	return out
}
