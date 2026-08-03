package orm

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

// SetupAdminParams holds the first company and administrator from the web setup wizard.
// No default admin/admin account is created; these values define the first user.
type SetupAdminParams struct {
	CompanyName string
	Lang        string
	FullName    string
	Email       string // used as login and stored in email
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

func allowedSetupLang(s string) bool {
	switch s {
	case "en_US", "en_GB", "fr_FR", "de_DE", "es_ES", "it_IT", "pt_BR", "nl_NL":
		return true
	default:
		return false
	}
}

// EnsureDefaultGroupsAndImplied creates kernel module categories, base security groups,
// xml ids base.group_system / base.group_user, and the Admin implies User edge.
// Call before module data sync (-i / -u) so addon XML can use ref('base.group_user') in implied_ids.
func EnsureDefaultGroupsAndImplied() error {
	ctx := ContextWithBypass(context.Background(), true)
	_, _, err := ensureDefaultKernelGroups(ctx)
	return err
}

func ensureDefaultKernelGroups(ctx context.Context) (adminGID int, userGID int, err error) {
	ctx = ContextWithBypass(ctx, true)
	if DB == nil {
		return 0, 0, nil
	}
	if err := EnsureSecurityJoinIndexes(); err != nil {
		return 0, 0, err
	}
	groupModel, ok := Registry["core.group"]
	if !ok || groupModel == nil {
		return 0, 0, fmt.Errorf("core.group not registered")
	}
	catModel, ok := Registry["sys.module.category"]
	if !ok || catModel == nil {
		return 0, 0, fmt.Errorf("sys.module.category not registered")
	}

	catAdminID, err := Upsert(ctx, catModel, map[string]interface{}{
		"name":     "Administration",
		"sequence": 1,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap sys.module.category Administration: %w", err)
	}
	catUserTypesID, err := Upsert(ctx, catModel, map[string]interface{}{
		"name":     "User types",
		"sequence": 2,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap sys.module.category User types: %w", err)
	}

	adminGID, err = Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "Administration / Settings",
		"category_id": catAdminID,
		"sequence":    1,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group admin: %w", err)
	}
	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module":  "base",
		"name":    "group_system",
		"model":   "core.group",
		"core_id": adminGID,
	}, "name")

	userGID, err = Upsert(ctx, groupModel, map[string]interface{}{
		"name":        "User types / Internal User",
		"category_id": catUserTypesID,
		"sequence":    10,
	}, "name")
	if err != nil {
		return 0, 0, fmt.Errorf("bootstrap core.group user: %w", err)
	}
	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module":  "base",
		"name":    "group_user",
		"model":   "core.group",
		"core_id": userGID,
	}, "name")

	_, _ = DB.ExecContext(ctx, `INSERT INTO `+GetTableName("core.group.implied")+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`, adminGID, userGID)
	return adminGID, userGID, nil
}

// EnsureBootstrapSecurityFromSetup creates groups, the first administrator, company, and ACLs.
// Call only from /setup/init after base module install when the database has no users yet.
func EnsureBootstrapSecurityFromSetup(p SetupAdminParams) error {
	if err := p.Validate(); err != nil {
		return err
	}
	return ensureBootstrapSecurity(context.Background(), &p)
}

// EnsureBootstrapSecurity ensures default groups and ACLs. If the database has no users yet,
// it returns an error so operators complete /setup instead of relying on a default account.
func EnsureBootstrapSecurity() error {
	return ensureBootstrapSecurity(context.Background(), nil)
}

func ensureBootstrapSecurity(ctx context.Context, first *SetupAdminParams) error {
	ctx = ContextWithBypass(ctx, true)
	if DB == nil {
		return nil
	}
	adminGID, userGID, err := ensureDefaultKernelGroups(ctx)
	if err != nil {
		return err
	}

	userModel, ok := Registry["core.user"]
	if !ok || userModel == nil {
		return fmt.Errorf("core.user not registered")
	}
	companyModel, ok := Registry["core.company"]
	if !ok || companyModel == nil {
		return fmt.Errorf("core.company not registered")
	}

	userTbl := GetTableName("core.user")
	var userCount int
	if err := DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+userTbl).Scan(&userCount); err != nil {
		return fmt.Errorf("count users: %w", err)
	}

	if userCount > 0 {
		ensureBootstrapACLs(ctx, adminGID, userGID)
		ensurePlatformDefaults(ctx)
		return nil
	}

	if first == nil {
		return fmt.Errorf("no users in this database: open /setup in your browser to create the administrator and finish initialization")
	}

	compID, err := Upsert(ctx, companyModel, map[string]interface{}{
		"name": first.CompanyName,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap company: %w", err)
	}
	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module":  "base",
		"name":    "main_company",
		"model":   "core.company",
		"core_id": compID,
	}, "name")

	login := strings.ToLower(first.Email)
	adminUID, err := Upsert(ctx, userModel, map[string]interface{}{
		"login":     login,
		"name":      first.FullName,
		"active":    true,
		"email":     first.Email,
		"lang":      first.Lang,
		"password":  "",
		"user_type": "internal",
	}, "login")
	if err != nil {
		return fmt.Errorf("bootstrap administrator: %w", err)
	}
	if adminUID == 0 {
		return fmt.Errorf("bootstrap administrator user id")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(first.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if _, err := DB.ExecContext(ctx, `UPDATE `+userTbl+` SET password = $1 WHERE id = $2`, string(hash), adminUID); err != nil {
		return fmt.Errorf("set administrator password: %w", err)
	}

	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module":  "base",
		"name":    "user_admin",
		"model":   "core.user",
		"core_id": adminUID,
	}, "name")

	if _, err := DB.ExecContext(ctx, `INSERT INTO `+GetTableName("core.group.user.rel")+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, adminGID); err != nil {
		return err
	}
	if _, err := DB.ExecContext(ctx, `INSERT INTO `+GetTableName("core.group.user.rel")+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, userGID); err != nil {
		return err
	}
	_, _ = DB.ExecContext(ctx, `UPDATE `+userTbl+` SET company_id = $1 WHERE id = $2`, compID, adminUID)

	ensureBootstrapACLs(ctx, adminGID, userGID)
	ensurePlatformDefaults(ctx)
	return nil
}

// ensurePlatformDefaults seeds config parameters and default sequences used by platform services.
func ensurePlatformDefaults(ctx context.Context) {
	_ = SetConfig(ctx, "auth.password_min_length", "8")
	if _, ok := Registry["sys.sequence"]; !ok {
		return
	}
	if _, err := SearchOne(ctx, "sys.sequence", map[string]interface{}{"code": "core.user.apikey"}); err == nil {
		return
	}
	inst := Registry["sys.sequence"]
	_, _ = Create(ctx, inst, map[string]interface{}{
		"name":        "API Key",
		"code":        "core.user.apikey",
		"prefix":      "KEY/",
		"padding":     4,
		"number_next": 1,
		"active":      true,
	})
}

func ensureBootstrapACLs(ctx context.Context, adminGID, userGID int) {
	installed, err := InstalledModuleNames(ctx)
	if err != nil {
		log.Printf("bootstrap ACL: installed modules: %v", err)
		installed = nil
	}
	for modelName := range Registry {
		if len(installed) > 0 && !ShouldMaterializeModel(modelName, installed) {
			continue
		}
		accName := fmt.Sprintf("access_%s_admin", strings.ReplaceAll(modelName, ".", "_"))
		if _, err := Upsert(ctx, SysAccess{}, map[string]interface{}{
			"name":        accName,
			"model":       modelName,
			"group_id":    NullableGroupIDForAccess(adminGID),
			"perm_read":   true,
			"perm_write":  true,
			"perm_create": true,
			"perm_unlink": true,
		}, "name"); err != nil {
			log.Printf("bootstrap ACL %s: %v", modelName, err)
		}
	}

	globalReads := []string{"sys.model_data", "sys.menu", "sys.action.window", "sys.view", "sys.module", "sys.module.category"}
	for _, m := range globalReads {
		accName := fmt.Sprintf("access_%s_global_read", strings.ReplaceAll(m, ".", "_"))
		_, _ = Upsert(ctx, SysAccess{}, map[string]interface{}{
			"name":        accName,
			"model":       m,
			"group_id":    nil,
			"perm_read":   true,
			"perm_write":  false,
			"perm_create": false,
			"perm_unlink": false,
		}, "name")
	}

	for _, pair := range []struct{ model, name string }{
		{"core.company", "access_core_company_user"},
		{"core.user", "access_core_user_user"},
		{"core.group", "access_core_group_user_read"},
		{"sys.access", "access_sys_access_user_read"},
		{"sys.rule", "access_sys_rule_user_read"},
	} {
		_, _ = Upsert(ctx, SysAccess{}, map[string]interface{}{
			"name":        pair.name,
			"model":       pair.model,
			"group_id":    NullableGroupIDForAccess(userGID),
			"perm_read":   true,
			"perm_write":  false,
			"perm_create": false,
			"perm_unlink": false,
		}, "name")
	}
}
