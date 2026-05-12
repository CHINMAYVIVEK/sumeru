package render

import (
	"strings"

	"sumeru/core/orm"
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

// EnrichShellPageData merges global shell branding and optional DB labels into page data.
func EnrichShellPageData(d *PageData) {
	d.AppName = AppDisplayName
	d.LogoURL = shell.LogoURL
	d.ShellCompany = strings.TrimSpace(shell.Company)
	d.ShellUser = strings.TrimSpace(shell.User)
	if orm.DB == nil {
		return
	}
	if d.ShellCompany == "" {
		if _, ok := orm.Registry["res.company"]; ok {
			tn := orm.GetTableName("res.company")
			var name string
			if err := orm.DB.QueryRow(`SELECT name FROM ` + tn + ` ORDER BY id ASC LIMIT 1`).Scan(&name); err == nil && strings.TrimSpace(name) != "" {
				d.ShellCompany = strings.TrimSpace(name)
			}
		}
	}
	if d.ShellUser == "" {
		if _, ok := orm.Registry["res.users"]; ok {
			tn := orm.GetTableName("res.users")
			var nm string
			q := `SELECT COALESCE(NULLIF(TRIM(name), ''), TRIM(login), '') FROM ` + tn + ` ORDER BY id ASC LIMIT 1`
			if err := orm.DB.QueryRow(q).Scan(&nm); err == nil && strings.TrimSpace(nm) != "" {
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
}
