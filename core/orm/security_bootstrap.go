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
	// Default groups
	adminGID, err := Upsert(ctx, ResGroups{}, map[string]interface{}{
		"name":     "Administration / Settings",
		"category": "Administration",
		"sequence": 1,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap res.groups admin: %w", err)
	}
	_, _ = Upsert(ctx, IrModelData{}, map[string]interface{}{
		"module": "user",
		"name":   "group_system",
		"model":  "res.groups",
		"res_id": adminGID,
	}, "name")

	userGID, err := Upsert(ctx, ResGroups{}, map[string]interface{}{
		"name":     "User types / Internal User",
		"category": "User types",
		"sequence": 10,
	}, "name")
	if err != nil {
		return fmt.Errorf("bootstrap res.groups user: %w", err)
	}
	_, _ = Upsert(ctx, IrModelData{}, map[string]interface{}{
		"module": "user",
		"name":   "group_user",
		"model":  "res.groups",
		"res_id": userGID,
	}, "name")

	// Admin implies internal user
	_, _ = DB.ExecContext(ctx, `INSERT INTO `+GetTableName("res.groups.implied")+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`, adminGID, userGID)

	// Admin user
	inst, ok := Registry["res.users"]
	if !ok || inst == nil {
		return fmt.Errorf("res.users not registered")
	}
	adminUID, err := Upsert(ctx, inst, map[string]interface{}{
		"login":    "admin",
		"name":     "Administrator",
		"active":   true,
		"email":    "admin@example.com",
		"lang":     "en_US",
		"password": "",
	}, "login")
	if err != nil {
		return fmt.Errorf("bootstrap res.users: %w", err)
	}
	if adminUID == 0 {
		return fmt.Errorf("bootstrap admin user id")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	row := DB.QueryRowContext(ctx, `SELECT COALESCE(NULLIF(TRIM(password), ''), '') FROM `+GetTableName("res.users")+` WHERE id = $1`, adminUID)
	var pw string
	_ = row.Scan(&pw)
	if pw == "" {
		if _, err := DB.ExecContext(ctx, `UPDATE `+GetTableName("res.users")+` SET password = $1 WHERE id = $2`, string(hash), adminUID); err != nil {
			return err
		}
		log.Printf("security bootstrap: set default password for login 'admin' (change after first login)")
	}

	_, _ = Upsert(ctx, IrModelData{}, map[string]interface{}{
		"module": "user",
		"name":   "user_admin",
		"model":  "res.users",
		"res_id": adminUID,
	}, "name")

	if _, err := DB.ExecContext(ctx, `INSERT INTO `+GetTableName("res.groups.user.rel")+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, adminUID, adminGID); err != nil {
		return err
	}

	// Full CRUD for Administration group on every registered model
	for modelName := range Registry {
		accName := fmt.Sprintf("access_%s_admin", strings.ReplaceAll(modelName, ".", "_"))
		if _, err := Upsert(ctx, IrModelAccess{}, map[string]interface{}{
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

	// Authenticated users may read UI metadata without belonging to Administration.
	globalReads := []string{"ir.model.data", "ir.ui.menu", "ir.actions.act_window", "ir.ui.view", "ir.module"}
	for _, m := range globalReads {
		accName := fmt.Sprintf("access_%s_global_read", strings.ReplaceAll(m, ".", "_"))
		_, _ = Upsert(ctx, IrModelAccess{}, map[string]interface{}{
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
		{"res.company", "access_res_company_user"},
		{"res.users", "access_res_users_user"},
		{"res.groups", "access_res_groups_user_read"},
		{"ir.model.access", "access_ir_model_access_user_read"},
		{"ir.rule", "access_ir_rule_user_read"},
	} {
		_, _ = Upsert(ctx, IrModelAccess{}, map[string]interface{}{
			"name":        pair.name,
			"model":       pair.model,
			"group_id":    NullableGroupIDForAccess(userGID),
			"perm_read":   true,
			"perm_write":  false,
			"perm_create": false,
			"perm_unlink": false,
		}, "name")
	}

	for _, m := range []string{"sale.order", "crm.lead", "product.product", "stock.picking", "mail.message"} {
		accName := fmt.Sprintf("access_%s_internal", strings.ReplaceAll(m, ".", "_"))
		_, _ = Upsert(ctx, IrModelAccess{}, map[string]interface{}{
			"name":        accName,
			"model":       m,
			"group_id":    NullableGroupIDForAccess(userGID),
			"perm_read":   true,
			"perm_write":  true,
			"perm_create": true,
			"perm_unlink": true,
		}, "name")
	}

	return nil
}
