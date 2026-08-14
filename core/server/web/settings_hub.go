package web

import (
	"fmt"
	"net/http"
	"strings"

	"sumeru/core/engine/render"
	"sumeru/core/orm"
)

type settingsHubLink struct {
	Name string
	Href string
}

// settingsHubSection groups links (e.g. Companies, Users) from the Settings menu tree.
type settingsHubSection struct {
	Title      string
	FilterText string
	Links      []settingsHubLink
}

// settingsHubAppTile is an alias for render.AppTile in settings hub templates.
type settingsHubAppTile = render.AppTile

type settingsHubData struct {
	Sections          []settingsHubSection
	AppTiles          []settingsHubAppTile
	CompaniesMenuHref string
}

// SettingsHubHandler renders the Settings overview at /web/settings.
// Requires base.group_user; admin-only sections (Localization, Users, etc.) need base.group_system
// and are omitted from the hub when the user lacks that group.
func SettingsHubHandler(w http.ResponseWriter, r *http.Request) {
	if !requireLogin(w, r) {
		return
	}
	ctx := r.Context()
	if !orm.UserHasGroupXML(ctx, orm.SecurityUID(ctx), "base.group_user") {
		http.Redirect(w, r, "/web/home", http.StatusFound)
		return
	}

	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_settings_root")
	if err != nil || rootID == 0 {
		http.Redirect(w, r, "/web/apps", http.StatusFound)
		return
	}
	menuIDStr := fmt.Sprintf("%d", rootID)

	_, sidebarMenus, _, _ := render.LoadShellMenus(ctx, menuIDStr)
	var sections []settingsHubSection
	for _, sec := range sidebarMenus {
		var links []settingsHubLink
		var ftParts []string
		ftParts = append(ftParts, strings.ToLower(strings.TrimSpace(sec.Name)))
		for _, sm := range sec.SubMenus {
			name := strings.TrimSpace(sm.Name)
			href := strings.TrimSpace(sm.Action)
			if name == "" || href == "" {
				continue
			}
			links = append(links, settingsHubLink{Name: name, Href: href})
			ftParts = append(ftParts, strings.ToLower(name))
		}
		if len(links) == 0 {
			continue
		}
		sections = append(sections, settingsHubSection{
			Title:      strings.TrimSpace(sec.Name),
			FilterText: strings.Join(ftParts, " "),
			Links:      links,
		})
	}

	raw, err := loadInstalledAppTiles(ctx, "")
	if err != nil {
		WebLogEvent(ctx, r.URL.Path, "Failed to list modules for settings hub", "load", "failure", err, nil)
		http.Error(w, "Failed to list modules", http.StatusInternalServerError)
		return
	}
	var appTiles []settingsHubAppTile
	for _, t := range raw {
		appTiles = append(appTiles, t)
	}

	companiesHref := ""
	if mid, _, err := orm.ResolveXmlId(ctx, "base.menu_company_companies"); err == nil && mid > 0 {
		companiesHref = fmt.Sprintf("/web?menu_id=%d", mid)
	}

	renderShellPage(w, r, shellPageOpts{
		Route:         r.URL.Path,
		InnerTemplate: "settings_hub_inner.html",
		InnerData: settingsHubData{
			Sections: sections, AppTiles: appTiles, CompaniesMenuHref: companiesHref,
		},
		MenuIDStr: menuIDStr,
		Page: render.PageData{
			Title:                "Settings",
			SettingsNavActive:    true,
			ActiveMenuID:         menuIDStr,
			SuppressActivityDock: true,
			BreadcrumbItems:      render.BuildSettingsHubBreadcrumbs(ctx),
			ViewStylesheetURLs:   []string{"/static/css/sumeru-settings-hub.css"},
			ExtraBodyClasses:     " sum-body--settings-hub",
		},
	})
}
