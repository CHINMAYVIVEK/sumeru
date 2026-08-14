package web

import (
	"net/http"
	"net/url"
	"strings"
)

const (
	appsRoute          = "/web/apps"
	appsPageTitle      = "Apps"
	appsInnerTemplate  = "apps_inner.html"
	appsModuleModel    = "sys.module"
	appsStylesheetURL  = "/static/css/sumeru-apps.css"
)

type appsNavVM struct {
	FilterAll, FilterInstalled, FilterUninstalled string
	ScopeAll, ScopeApps, ScopeTechnical           string
}

// appsBrowseState holds normalized Apps page query parameters.
type appsBrowseState struct {
	Message     string
	Layout      string
	Filter      string
	Scope       string
	SearchQuery string
	ModuleName  string
	Editing     bool
}

func parseAppsBrowseState(r *http.Request) appsBrowseState {
	return appsBrowseState{
		Message:     strings.TrimSpace(r.URL.Query().Get("msg")),
		Layout:      layoutFromQuery(r),
		ModuleName:  strings.TrimSpace(r.URL.Query().Get("module")),
		Editing:     strings.TrimSpace(r.URL.Query().Get("edit")) == "1",
		Filter:      normalizeAppsFilter(r.URL.Query().Get("filter")),
		Scope:       normalizeAppsScope(r.URL.Query().Get("scope")),
		SearchQuery: strings.TrimSpace(r.URL.Query().Get("q")),
	}
}

// parseAppsBrowseStateFromForm reads apps_* POST fields preserved across module action redirects.
func parseAppsBrowseStateFromForm(r *http.Request) appsBrowseState {
	return appsBrowseState{
		Layout:      layoutFromForm(r, appsLayoutField),
		Filter:      normalizeAppsFilter(r.FormValue(appsFilterField)),
		Scope:       normalizeAppsScope(r.FormValue(appsScopeField)),
		SearchQuery: strings.TrimSpace(r.FormValue(appsSearchField)),
	}
}

// appsRedirectURL builds an Apps page URL with optional flash message and browse filters.
func appsRedirectURL(message string, browse appsBrowseState) string {
	query := url.Values{}
	if strings.TrimSpace(message) != "" {
		query.Set("msg", strings.TrimSpace(message))
	}
	appendAppsQueryBase(query, browse.Layout, browse.Filter, browse.Scope, browse.SearchQuery)
	if encoded := query.Encode(); encoded != "" {
		return appsRoute + "?" + encoded
	}
	return appsRoute
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

// filterAppsModulesByBrowse applies search/filter and splits by application vs technical scope.
func filterAppsModulesByBrowse(modules []appsModule, browse appsBrowseState) (appModules, techModules []appsModule) {
	for _, moduleEntry := range modules {
		if !appsModuleMatchesSearch(moduleEntry, browse.SearchQuery) || !appsModuleMatchesFilter(moduleEntry, browse.Filter) {
			continue
		}
		if moduleEntry.Application {
			appModules = append(appModules, moduleEntry)
		} else {
			techModules = append(techModules, moduleEntry)
		}
	}
	switch browse.Scope {
	case "apps":
		techModules = nil
	case "technical":
		appModules = nil
	}
	return appModules, techModules
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

// appsDetailURL builds /web/apps?… links for module detail, edit, and cancel views.
func appsDetailURL(layout, filter, scope, searchQuery, moduleName string, editing bool) string {
	query := url.Values{}
	appendAppsQueryBase(query, layout, filter, scope, searchQuery)
	query.Set("module", moduleName)
	if editing {
		query.Set("edit", "1")
	}
	return appsRoute + "?" + query.Encode()
}

func appsLink(layout, filter, scope, q string) string {
	v := url.Values{}
	appendAppsQueryBase(v, layout, filter, scope, q)
	return appsRoute + "?" + v.Encode()
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
