package orm

import (
	"fmt"
	"strings"
)

// EnsureSecurityJoinIndexes creates composite unique indexes for M2M-style tables.
func EnsureSecurityJoinIndexes() error {
	if DB == nil {
		return nil
	}
	// 1. Ensure Join Tables exist
	joinTables := []string{
		`CREATE TABLE IF NOT EXISTS ` + GetTableName("core.group.user.rel") + ` (user_id BIGINT NOT NULL, group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + GetTableName("core.group.implied") + ` (group_id BIGINT NOT NULL, implied_group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + GetTableName("sys.rule.group.rel") + ` (rule_id BIGINT NOT NULL, group_id BIGINT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS ` + GetTableName("core.user.company.rel") + ` (user_id BIGINT NOT NULL, company_id BIGINT NOT NULL)`,
	}
	for _, q := range joinTables {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("create join table: %w", err)
		}
	}

	// 2. Ensure Composite Unique Indexes
	stmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS core_group_users_rel_user_group_uq ON ` + GetTableName("core.group.user.rel") + ` (user_id, group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS core_group_implied_gid_hid_uq ON ` + GetTableName("core.group.implied") + ` (group_id, implied_group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS sys_rule_group_rel_rule_group_uq ON ` + GetTableName("sys.rule.group.rel") + ` (rule_id, group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS core_user_company_rel_uq ON ` + GetTableName("core.user.company.rel") + ` (user_id, company_id)`,
	}
	for _, q := range stmts {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return nil
}

// NullableGroupIDForAccess returns SQL argument for sys.access.group_id (NULL if 0).
func NullableGroupIDForAccess(groupID int) interface{} {
	if groupID <= 0 {
		return nil
	}
	return int64(groupID)
}

// NormalizeAccessGroupList splits comma-separated XML ids.
func NormalizeAccessGroupList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
