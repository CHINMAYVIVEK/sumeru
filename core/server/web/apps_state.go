package web

import (
	"fmt"
	"net/url"
	"strings"
)

type appsNavVM struct {
	FilterAll, FilterInstalled, FilterUninstalled string
	ScopeAll, ScopeApps, ScopeTechnical           string
}

func stringField(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case nil:
		return ""
	default:
		return fmt.Sprint(t)
	}
}

func boolField(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case int64:
		return t != 0
	case []byte:
		return string(t) == "t" || string(t) == "true" || string(t) == "1"
	default:
		s := strings.ToLower(stringField(v))
		return s == "true" || s == "t" || s == "1"
	}
}

func normalizeAppsFilter(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "installed", "uninstalled":
		return strings.ToLower(strings.TrimSpace(s))
	default:
		return "all"
	}
}

func normalizeAppsScope(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "apps", "technical":
		return strings.ToLower(strings.TrimSpace(s))
	default:
		return "all"
	}
}

func appsModuleMatchesFilter(m appsModule, filter string) bool {
	switch filter {
	case "installed":
		return m.State == "installed"
	case "uninstalled":
		return m.State != "installed"
	default:
		return true
	}
}

func appsModuleMatchesSearch(m appsModule, q string) bool {
	q = strings.TrimSpace(q)
	if q == "" {
		return true
	}
	lq := strings.ToLower(q)
	return strings.Contains(strings.ToLower(m.Name), lq) || strings.Contains(strings.ToLower(m.DisplayName), lq)
}

func appendAppsQueryBase(v url.Values, layout, filter, scope, q string) {
	if layout == "list" || layout == "grid" {
		v.Set("layout", layout)
	}
	if filter != "" && filter != "all" {
		v.Set("filter", filter)
	}
	if scope != "" && scope != "all" {
		v.Set("scope", scope)
	}
	if strings.TrimSpace(q) != "" {
		v.Set("q", strings.TrimSpace(q))
	}
}

func appsBrowseQuery(layout, filter, scope, q string) string {
	v := url.Values{}
	appendAppsQueryBase(v, layout, filter, scope, q)
	return v.Encode()
}

func appsLink(layout, filter, scope, q string) string {
	v := url.Values{}
	appendAppsQueryBase(v, layout, filter, scope, q)
	return "/web/apps?" + v.Encode()
}

func buildAppsNavVM(layout, filter, scope, q string) appsNavVM {
	return appsNavVM{
		FilterAll:         appsLink(layout, "all", scope, q),
		FilterInstalled:   appsLink(layout, "installed", scope, q),
		FilterUninstalled: appsLink(layout, "uninstalled", scope, q),
		ScopeAll:          appsLink(layout, filter, "all", q),
		ScopeApps:         appsLink(layout, filter, "apps", q),
		ScopeTechnical:    appsLink(layout, filter, "technical", q),
	}
}
