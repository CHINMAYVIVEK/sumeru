package web

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"sumeru/core/engine/parser"
	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

const workspaceListSearchParam = render.WorkspaceSearchParam

func listSearchQuery(r *http.Request) string {
	return strings.TrimSpace(r.URL.Query().Get(workspaceListSearchParam))
}

func listSearchFieldNames(views ...*parser.View) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, view := range views {
		if view == nil {
			continue
		}
		for _, f := range view.Field {
			n := strings.TrimSpace(f.Name)
			if n == "" {
				continue
			}
			if _, ok := seen[n]; ok {
				continue
			}
			seen[n] = struct{}{}
			out = append(out, n)
		}
	}
	return out
}

func workspaceListDomain(ctx context.Context, actionData map[string]interface{}, view, searchView *parser.View, searchQuery, filterCSV string) [][]interface{} {
	base := actionListDomain(ctx, actionData)
	model := ""
	if view != nil {
		model = view.Model
	}
	searchFields := listSearchFieldNames(view, searchView)
	if searchQuery != "" && model != "" {
		search := orm.BuildListSearchDomain(model, searchFields, searchQuery)
		base = orm.MergeDomains(base, search)
	}
	uid := orm.SecurityUID(ctx)
	for _, name := range splitCommaSeparatedValues(filterCSV) {
		f := findSearchFilter(searchView, name)
		if f == nil || strings.TrimSpace(f.Domain) == "" {
			continue
		}
		dom, err := orm.ParseDomainJSON(f.Domain)
		if err != nil || len(dom) == 0 {
			continue
		}
		dom = orm.ResolveDomainXMLRefs(ctx, dom)
		dom = orm.SubstituteDomainUID(dom, uid)
		base = orm.MergeDomains(base, dom)
	}
	return base
}

func findSearchFilter(searchView *parser.View, name string) *parser.SearchFilter {
	if searchView == nil {
		return nil
	}
	name = strings.TrimSpace(name)
	for i := range searchView.SearchFilter {
		if searchView.SearchFilter[i].Name == name {
			return &searchView.SearchFilter[i]
		}
	}
	return nil
}

func workspaceListSearchURL(req workspaceRequest) string {
	offset := ""
	if req.listOffset > 0 {
		offset = strconv.Itoa(req.listOffset)
	}
	return render.WorkspaceURL(render.WorkspaceQuery{
		ActionID: req.actionID,
		MenuID:   req.menuID,
		ViewType: workspaceViewModeList,
		Search:   req.listSearch,
		Model:    req.model,
		Filter:   req.listFilter,
		Sort:     req.listSort,
		Offset:   offset,
		GroupBy:  req.listGroupBy,
	})
}
