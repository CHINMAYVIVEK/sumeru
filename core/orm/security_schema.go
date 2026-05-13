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
	stmts := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS res_groups_users_rel_user_group_uq ON ` + GetTableName("res.groups.user.rel") + ` (user_id, group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS res_groups_implied_gid_hid_uq ON ` + GetTableName("res.groups.implied") + ` (group_id, implied_group_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ir_rule_group_rel_rule_group_uq ON ` + GetTableName("ir.rule.group.rel") + ` (rule_id, group_id)`,
	}
	for _, q := range stmts {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("%s: %w", q, err)
		}
	}
	return nil
}

// NullableGroupIDForAccess returns SQL argument for ir.model.access.group_id (NULL if 0).
func NullableGroupIDForAccess(groupID int) interface{} {
	if groupID <= 0 {
		return nil
	}
	return int64(groupID)
}

// CoalesceGroupID scans sql.NullInt64 style into int (0 = none / global row semantics).
func CoalesceGroupID(v interface{}) int {
	if v == nil {
		return 0
	}
	n, ok := CoerceInt64(v)
	if !ok || n == 0 {
		return 0
	}
	return int(n)
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
