package render

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"sumeru/core/mail"
	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// AppDisplayName is the product name shown in the shell, document title, and manifests.
const AppDisplayName = "Sumeru"

// ShellBranding is header chrome (logo + labels) set once at process start from config.
type ShellBranding struct {
	LogoURL string // e.g. "/static/app-logo" or empty
	Company string
	User    string
}

var shell ShellBranding

// SetShellBranding configures header logo URL and optional company / user labels.
func SetShellBranding(b ShellBranding) {
	shell = b
}

// ShellLogoURL returns the configured shell logo URL (e.g. "/static/app-logo"), or empty.
func ShellLogoURL() string {
	return strings.TrimSpace(shell.LogoURL)
}

// EnrichShellPageData merges global shell branding and optional DB labels into page data.
func EnrichShellPageData(ctx context.Context, d *PageData) {
	d.AppName = AppDisplayName
	d.LogoURL = shell.LogoURL
	d.ShellCompany = strings.TrimSpace(shell.Company)
	d.ShellUser = strings.TrimSpace(shell.User)
	if orm.DB == nil {
		return
	}
	if d.ShellCompany == "" {
		if _, ok := orm.Registry["core.company"]; ok {
			tn := orm.GetTableName("core.company")
			var name string
			if err := orm.DB.QueryRowContext(ctx, `SELECT name FROM `+tn+` ORDER BY id ASC LIMIT 1`).Scan(&name); err == nil && strings.TrimSpace(name) != "" {
				d.ShellCompany = strings.TrimSpace(name)
			}
		}
	}
	if d.ShellUser == "" {
		if uid := orm.UIDFromContext(ctx); uid > 0 {
			if u, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": uid}); err == nil {
				d.ShellUser = strings.TrimSpace(orm.AsString(u["name"]))
				if d.ShellUser == "" {
					d.ShellUser = strings.TrimSpace(orm.AsString(u["login"]))
				}
			}
		}
	}
	// Fallback: only in dev mode — avoids leaking the admin name to unauthenticated users.
	if d.ShellUser == "" && config.AppConfig.DevMode {
		if _, ok := orm.Registry["core.user"]; ok {
			tn := orm.GetTableName("core.user")
			var nm string
			q := `SELECT COALESCE(NULLIF(TRIM(name), ''), TRIM(login), '') FROM ` + tn + ` ORDER BY id ASC LIMIT 1`
			if err := orm.DB.QueryRowContext(ctx, q).Scan(&nm); err == nil && strings.TrimSpace(nm) != "" {
				d.ShellUser = strings.TrimSpace(nm)
			}
		}
	}
	if d.UserInitial == "" && d.ShellUser != "" {
		r := []rune(d.ShellUser)
		if len(r) > 0 {
			d.UserInitial = strings.ToUpper(string(r[0]))
		}
	}

	d.ActivityEnabled = mail.CompanyChatterEnabled(ctx) && mail.CompanyActivityPanelEnabled(ctx)
	if d.SuppressActivityDock {
		d.ActivityEnabled = false
	}
	if !d.ActivityEnabled {
		d.ActivityLogItems = nil
		return
	}
	rows, err := mail.QueryActivityLog(ctx, 40, d.ActivityContextModel, d.ActivityContextRecordID)
	if err != nil {
		d.ActivityLogItems = nil
		return
	}
	for _, r := range rows {
		author := strings.TrimSpace(r.Author)
		if author == "" {
			author = "System"
		}
		meta := author
		if !r.CreateDate.IsZero() {
			meta = fmt.Sprintf("%s · %s", author, shortRelTime(r.CreateDate))
		}
		d.ActivityLogItems = append(d.ActivityLogItems, ActivityItem{Meta: meta, Body: strings.TrimSpace(r.Body)})
	}

	// EXECUTE SHELL HOOKS (e.g. AI Assistant)
	var shellExtra strings.Builder
	for _, hook := range ShellHooks {
		shellExtra.WriteString(string(hook(ctx, nil, false)))
	}
	d.ShellExtraHTML = template.HTML(shellExtra.String())
}

func shortRelTime(t time.Time) string {
	t = t.UTC()
	now := time.Now().UTC()
	if t.After(now) {
		t = now
	}
	d := now.Sub(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 48*time.Hour:
		return "yesterday"
	default:
		return t.Local().Format("Jan 02")
	}
}
