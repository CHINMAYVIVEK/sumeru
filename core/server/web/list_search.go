package web

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/orm"
)

const workspaceListSearchParam = "q"

func listSearchQuery(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get(workspaceListSearchParam))
}

func listSearchFieldNames(view *parser.View) []string {
	if view == nil {
		return nil
	}
	out := make([]string, 0, len(view.Field))
	for _, f := range view.Field {
		if n := strings.TrimSpace(f.Name); n != "" {
			out = append(out, n)
		}
	}
	return out
}

func workspaceListDomain(ctx context.Context, actionData map[string]interface{}, view *parser.View, searchQuery string) [][]interface{} {
	base := actionListDomain(ctx, actionData)
	if searchQuery == "" || view == nil {
		return base
	}
	search := orm.BuildListSearchDomain(view.Model, listSearchFieldNames(view), searchQuery)
	return orm.MergeDomains(base, search)
}

func workspaceListSearchURL(actionID int, menuID, searchQuery string) string {
	vals := url.Values{}
	if actionID > 0 {
		vals.Set(workspaceActionParam, strconv.Itoa(actionID))
	}
	if menuID = strings.TrimSpace(menuID); menuID != "" {
		vals.Set(workspaceMenuIDParam, menuID)
	}
	vals.Set(workspaceViewTypeParam, workspaceViewModeList)
	if searchQuery = strings.TrimSpace(searchQuery); searchQuery != "" {
		vals.Set(workspaceListSearchParam, searchQuery)
	}
	encoded := vals.Encode()
	if encoded == "" {
		return workspaceRoute
	}
	return workspaceRoute + "?" + encoded
}
