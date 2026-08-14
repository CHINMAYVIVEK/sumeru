package orm

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// SetupAdminParams holds the first company and administrator from the web setup wizard.
type SetupAdminParams struct {
	CompanyName string
	Lang        string
	FullName    string
	Email       string
	Password    string
}

// Validate normalizes fields and checks constraints for the setup wizard.
func (p *SetupAdminParams) Validate() error {
	p.CompanyName = strings.TrimSpace(p.CompanyName)
	p.FullName = strings.TrimSpace(p.FullName)
	p.Email = strings.TrimSpace(p.Email)
	p.Password = strings.TrimSpace(p.Password)
	p.Lang = strings.TrimSpace(p.Lang)
	if p.Lang == "" {
		p.Lang = "en_US"
	}
	if len(p.CompanyName) < 1 || len(p.CompanyName) > 200 {
		return fmt.Errorf("company name must be between 1 and 200 characters")
	}
	if len(p.FullName) < 1 || len(p.FullName) > 200 {
		return fmt.Errorf("full name is required")
	}
	if len(p.Email) < 3 || len(p.Email) > 254 || !strings.Contains(p.Email, "@") {
		return fmt.Errorf("enter a valid email address (used as your login)")
	}
	if n := utf8.RuneCountInString(p.Password); n < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(p.Password) > 72 {
		return fmt.Errorf("password must be at most 72 bytes for storage")
	}
	if !allowedSetupLang(p.Lang) {
		return fmt.Errorf("unsupported language")
	}
	return nil
}

func allowedSetupLang(lang string) bool {
	switch lang {
	case "en_US", "en_GB", "fr_FR", "de_DE", "es_ES", "it_IT", "pt_BR", "nl_NL":
		return true
	default:
		return false
	}
}
