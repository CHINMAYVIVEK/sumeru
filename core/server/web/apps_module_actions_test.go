package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppsRedirectURL(t *testing.T) {
	browse := appsBrowseState{
		Layout:      "grid",
		Filter:      "installed",
		Scope:       "apps",
		SearchQuery: "crm",
	}
	got := appsRedirectURL("installed_sale", browse)
	assertQueryContains(t, got, map[string]string{
		"msg":    "installed_sale",
		"layout": "grid",
		"filter": "installed",
		"scope":  "apps",
		"q":      "crm",
	})

	empty := appsRedirectURL("", appsBrowseState{})
	if empty != appsRoute {
		t.Fatalf("got %q want %q", empty, appsRoute)
	}
}

func TestParseAppsBrowseStateFromForm(t *testing.T) {
	body := "apps_layout=list&apps_filter=installed&apps_scope=technical&apps_q=sale"
	req := httptest.NewRequest("POST", moduleActionRoute, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := req.ParseForm(); err != nil {
		t.Fatal(err)
	}
	browse := parseAppsBrowseStateFromForm(req)
	if browse.Layout != "list" || browse.Filter != "installed" || browse.Scope != "technical" || browse.SearchQuery != "sale" {
		t.Fatalf("unexpected browse: %+v", browse)
	}
}

func TestAppsDetailRedirectURL(t *testing.T) {
	browse := appsBrowseState{
		Layout:      "grid",
		Filter:      "all",
		Scope:       "all",
		SearchQuery: "",
		ModuleName:  "sale",
	}
	got := appsDetailRedirectURL("saved", browse)
	assertQueryContains(t, got, map[string]string{
		"msg":    "saved",
		"module": "sale",
		"layout": "grid",
	})
}
