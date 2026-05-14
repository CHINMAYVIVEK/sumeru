package orm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// EnsureBootstrapSecurity creates default groups, admin user, ACLs, and implied links when missing.
func EnsureBootstrapSecurity() error {
	ctx := ContextWithBypass(context.Background(), true)
	if DB == nil {
		return nil
	}
	if err := EnsureSecurityJoinIndexes(); err != nil {
		return err
	}

	// 1. Resolve Core Models from Registry
	groupModel, ok := Registry["core.group"]
	if !ok || groupModel == nil {
		return fmt.Errorf("core.group not registered")
	}
	userModel, ok := Registry["core.user"]
	if !ok || userModel == nil {
		return fmt.Errorf("core.user not registered")
	}
	companyModel, ok := Registry["core.company"]
	if !ok || companyModel == nil {
		return fmt.Errorf("core.company not registered")
	}

	// 2. Default groups
	adminGID, err := Upsert(ctx, groupModel, map[string]interface{}{
		"name":     "Administration / Settings",
		"category": "Administration",
		"sequence": 1,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap core.group admin: %w", err)
	}
	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module": "base",
		"name":   "group_system",
		"model":  "core.group",
		"core_id": adminGID,
	}, "name")

	userGID, err := Upsert(ctx, groupModel, map[string]interface{}{
		"name":     "User types / Internal User",
		"category": "User types",
		"sequence": 10,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap core.group user: %w", err)
	}
	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module": "base",
		"name":   "group_user",
		"model":  "core.group",
		"core_id": userGID,
	}, "name")

	// Admin implies internal user
	_, _ = DB.ExecContext(ctx, `INSERT INTO `+GetTableName("core.group.implied")+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`, adminGID, userGID)

	// 3. Admin user
	adminUID, err := Upsert(ctx, userModel, map[string]interface{}{
		"login":    "admin",
		"name":     "Administrator",
		"active":   true,
		"email":    "admin@example.com",
		"lang":     "en_US",
		"password": "",
	}, "login")
	if err != nil {
		return fmt.Errorf("bootstrap core.user: %w", err)
	}
	if adminUID == 0 {
		return fmt.Errorf("bootstrap admin user id")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	row := DB.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(TRIM(password), ''), '') FROM `+GetTableName("core.user")+` WHERE id = $1`, adminUID)
	var pw string
	_ = row.Scan(&pw)
	if pw == "" {
		if _, err := DB.ExecContext(ctx, `UPDATE `+GetTableName("core.user")+` SET password = $1 WHERE id = $2`, string(hash), adminUID); err != nil {
			return err
		}
		log.Printf("security bootstrap: set default password for login 'admin' (change after first login)")
	}

	_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
		"module": "base",
		"name":   "user_admin",
		"model":  "core.user",
		"core_id": adminUID,
	}, "name")

	if _, err := DB.ExecContext(ctx, `INSERT INTO `+GetTableName("core.group.user.rel")+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, adminGID); err != nil {
		return err
	}

	// 4. Default Company
	compID, err := Upsert(ctx, companyModel, map[string]interface{}{
		"name": "My Company",
	}, "name")
	if err == nil {
		_, _ = Upsert(ctx, SysModelData{}, map[string]interface{}{
			"module": "base",
			"name":   "main_company",
			"model":  "core.company",
			"core_id": compID,
		}, "name")
		// Link admin to company
		_, _ = DB.ExecContext(ctx, `UPDATE `+GetTableName("core.user")+` SET company_id = $1 WHERE id = $2`, compID, adminUID)
	}

	// 5. Full CRUD for Administration group on every registered model
	for modelName := range Registry {
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

	// 6. Global Read Access for Metadata
	globalReads := []string{"sys.model_data", "sys.menu", "sys.action.window", "sys.view", "sys.module"}
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

	// 7. Core User Permissions
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

	return nil
}
