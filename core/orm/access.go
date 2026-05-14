package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const superuserUID = 1

// EffectiveGroupIDs returns all group ids for uid (direct + implied). Superuser gets all group ids.
func EffectiveGroupIDs(ctx context.Context, uid int) ([]int, error) {
	if uid <= 0 || DB == nil {
		return nil, nil
	}
	if uid == superuserUID {
		rows, err := DB.QueryContext(ctx, `SELECT id FROM `+GetTableName("core.group")+` ORDER BY id`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var all []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return nil, err
			}
			all = append(all, id)
		}
		return all, rows.Err()
	}
	rows, err := DB.QueryContext(ctx,
		`SELECT group_id FROM `+GetTableName("core.group.user.rel")+` WHERE user_id = $1`,
		uid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	direct := map[int]struct{}{}
	for rows.Next() {
		var gid int
		if err := rows.Scan(&gid); err != nil {
			return nil, err
		}
		direct[gid] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Expand implied groups (BFS).
	out := map[int]struct{}{}
	queue := make([]int, 0, len(direct))
	for g := range direct {
		out[g] = struct{}{}
		queue = append(queue, g)
	}
	implTbl := GetTableName("core.group.implied")
	for len(queue) > 0 {
		gid := queue[0]
		queue = queue[1:]
		r2, err := DB.QueryContext(ctx, `SELECT implied_group_id FROM `+implTbl+` WHERE group_id = $1`, gid)
		if err != nil {
			return nil, err
		}
		for r2.Next() {
			var hid int
			if err := r2.Scan(&hid); err != nil {
				r2.Close()
				return nil, err
			}
			if _, ok := out[hid]; !ok {
				out[hid] = struct{}{}
				queue = append(queue, hid)
			}
		}
		if err := r2.Err(); err != nil {
			r2.Close()
			return nil, err
		}
		r2.Close()
	}
	var list []int
	for g := range out {
		list = append(list, g)
	}
	return list, nil
}

func intSliceContains(haystack []int, needle int) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

// CheckModelAccess verifies uid may perform op on model (read|write|create|unlink).
func CheckModelAccess(ctx context.Context, uid int, model string, op string) error {
	if SecurityBypass(ctx) {
		return nil
	}
	if uid == superuserUID {
		return nil
	}
	if _, ok := Registry[model]; !ok {
		return fmt.Errorf("unknown model %q", model)
	}
	if uid <= 0 {
		return fmt.Errorf("access denied on %s: not authenticated", model)
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}
	want := permColumnForOp(op)
	if want == "" {
		return fmt.Errorf("unknown operation %q", op)
	}
	accTbl := GetTableName("sys.access")
	q := `SELECT group_id, ` + want + ` FROM ` + accTbl + ` WHERE model = $1 AND ` + want + ` = true`
	rows, err := DB.QueryContext(ctx, q, model)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var gid sql.NullInt64
		var perm bool
		if err := rows.Scan(&gid, &perm); err != nil {
			return err
		}
		if !perm {
			continue
		}
		if !gid.Valid || gid.Int64 == 0 {
			return nil
		}
		if intSliceContains(groups, int(gid.Int64)) {
			return nil
		}
	}
	return fmt.Errorf("access denied on %s for operation %s", model, op)
}

func permColumnForOp(op string) string {
	switch strings.ToLower(strings.TrimSpace(op)) {
	case "read":
		return "perm_read"
	case "write":
		return "perm_write"
	case "create":
		return "perm_create"
	case "unlink":
		return "perm_unlink"
	default:
		return ""
	}
}

// ApplicableRuleDomains returns parsed domains for rules that apply to uid on model for op.
// Implements Sumeru-style logic: (Global rules ANDed) AND (Group rules ORed together).
func ApplicableRuleDomains(ctx context.Context, uid int, model string, op string) ([][][]interface{}, error) {
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil, nil
	}
	if uid <= 0 {
		return nil, nil
	}
	col := "perm_" + strings.ToLower(strings.TrimSpace(op))
	if col != "perm_read" && col != "perm_write" && col != "perm_create" && col != "perm_unlink" {
		col = "perm_read"
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	ruleTbl := GetTableName("sys.rule")
	relTbl := GetTableName("sys.rule.group.rel")
	q := `SELECT r.id, r.domain_force, r.active, r.` + col + ` FROM ` + ruleTbl + ` r WHERE r.model = $1 AND r.active = true AND r.` + col + ` = true`
	rows, err := DB.QueryContext(ctx, q, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var globalDomains [][][]interface{}
	var groupDomains [][][]interface{}

	for rows.Next() {
		var id int
		var domainForce string
		var active, permOp bool
		if err := rows.Scan(&id, &domainForce, &active, &permOp); err != nil {
			return nil, err
		}
		if !active || !permOp {
			continue
		}
		var n int
		err := DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM `+relTbl+` WHERE rule_id = $1`, id).Scan(&n)
		if err != nil {
			return nil, err
		}

		dom, err := ParseDomainJSON(domainForce)
		if err != nil {
			return nil, fmt.Errorf("rule %d: %w", id, err)
		}
		dom = SubstituteDomainUID(dom, uid)

		if n == 0 {
			// Global rule (no groups)
			if len(dom) > 0 {
				globalDomains = append(globalDomains, dom)
			}
		} else {
			// Group rule
			r2, err := DB.QueryContext(ctx, `SELECT group_id FROM `+relTbl+` WHERE rule_id = $1`, id)
			if err != nil {
				return nil, err
			}
			match := false
			for r2.Next() {
				var gid int
				if err := r2.Scan(&gid); err != nil {
					r2.Close()
					return nil, err
				}
				if intSliceContains(groups, gid) {
					match = true
					break
				}
			}
			r2.Close()
			if match && len(dom) > 0 {
				groupDomains = append(groupDomains, dom)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Sumeru logic: (Global1 AND Global2 ...) AND (GroupRule1 OR GroupRule2 ...)
	// Global domains are returned individually so callers AND all of them.
	// Group domains are OR-merged into a single domain using the "|" prefix operator.
	out := append([][][]interface{}(nil), globalDomains...)

	switch len(groupDomains) {
	case 0:
		// No group rules applicable — no additional restriction.
	case 1:
		// Single group rule — add directly (OR prefix not needed).
		out = append(out, groupDomains[0])
	default:
		// Multiple group rules — OR-merge them.
		// Sumeru prefix-Polish OR: prepend (N-1) "|" operators then all N leaf conditions.
		merged := [][]interface{}{}
		for i := 0; i < len(groupDomains)-1; i++ {
			merged = append(merged, []interface{}{"|"})
		}
		for _, gd := range groupDomains {
			merged = append(merged, gd...)
		}
		out = append(out, merged)
	}

	return out, nil
}

// MergeRuleDomainsIntoSearch AND-merges rule domains into the search domain.
func MergeRuleDomainsIntoSearch(ctx context.Context, uid int, model string, op string, base [][]interface{}) ([][]interface{}, error) {
	ruleParts, err := ApplicableRuleDomains(ctx, uid, model, op)
	if err != nil {
		return nil, err
	}
	out := append([][]interface{}(nil), base...)
	for _, p := range ruleParts {
		out = append(out, p...)
	}
	return out, nil
}

// CheckRecordRules verify record satisfies all applicable rules.
func CheckRecordRules(ctx context.Context, uid int, model string, op string, rec map[string]interface{}) error {
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if uid <= 0 {
		return fmt.Errorf("access denied")
	}
	parts, err := ApplicableRuleDomains(ctx, uid, model, op)
	if err != nil {
		return err
	}
	for _, dom := range parts {
		if !RecordMatchesDomain(rec, dom) {
			return fmt.Errorf("record rule failed for model %s", model)
		}
	}
	return nil
}

// UserHasAnyAccessGroup resolves comma-separated XML ids; returns true if accessGroups empty or user matches.
func UserHasAnyAccessGroup(ctx context.Context, uid int, accessGroupsCSV string) bool {
	if SecurityBypass(ctx) || uid == superuserUID {
		return true
	}
	ids := NormalizeAccessGroupList(accessGroupsCSV)
	if len(ids) == 0 {
		return true
	}
	if uid <= 0 {
		return false
	}
	eff, err := EffectiveGroupIDs(ctx, uid)
	if err != nil || len(eff) == 0 {
		return false
	}
	for _, xmlID := range ids {
		gid, _, err := ResolveXmlId(ctx, xmlID)
		if err != nil || gid == 0 {
			continue
		}
		if intSliceContains(eff, gid) {
			return true
		}
	}
	return false
}

// CheckStageApproval verifies if uid has permission to move record to targetState.
func CheckStageApproval(ctx context.Context, model string, id int, targetState string) error {
	if SecurityBypass(ctx) || SecurityUID(ctx) == superuserUID {
		return nil
	}
	uid := SecurityUID(ctx)
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}
	
	// Get current state to check from_state rules
	before, err := SearchOne(ctx, model, map[string]interface{}{"id": id})
	if err != nil {
		// If record not found, we can't check from_state. For now, assume it's okay (e.g. create case handled elsewhere).
		return nil
	}
	currentState := AsString(before["state"])

	appTbl := GetTableName("sys.approval_rule")
	// Find rules for this model and target state
	q := `SELECT group_id, COALESCE(from_state, '') FROM ` + appTbl + ` WHERE model = $1 AND to_state = $2 AND require_approval = true`
	rows, err := DB.QueryContext(ctx, q, model, targetState)
	if err != nil {
		// If table doesn't exist yet or other error, allow.
		return nil
	}
	defer rows.Close()
	
	hasRule := false
	match := false
	for rows.Next() {
		hasRule = true
		var gid int
		var fromState string
		if err := rows.Scan(&gid, &fromState); err != nil {
			return err
		}
		
		// If fromState is specified, it must match current state
		if fromState != "" && fromState != currentState {
			continue
		}

		if intSliceContains(groups, gid) {
			match = true
			break
		}
	}
	
	if hasRule && !match {
		return fmt.Errorf("approval required for transition to state %q (from %q)", targetState, currentState)
	}
	return nil
}

// SetUserGroupLinks replaces group membership for a user (explicit rel rows only).
func SetUserGroupLinks(ctx context.Context, userID int, groupIDs []int) error {
	if DB == nil {
		return fmt.Errorf("no database")
	}
	tbl := GetTableName("core.group.user.rel")
	if _, err := DB.ExecContext(ctx, `DELETE FROM `+tbl+` WHERE user_id = $1`, userID); err != nil {
		return err
	}
	for _, gid := range groupIDs {
		if gid <= 0 {
			continue
		}
		if _, err := DB.ExecContext(ctx, `INSERT INTO `+tbl+` (user_id, group_id) VALUES ($1, $2) ON CONFLICT (user_id, group_id) DO NOTHING`, userID, gid); err != nil {
				return err
			}
	}
	return nil
}

// ListAllGroupRows returns id,name for UI pickers.
func ListAllGroupRows(ctx context.Context) ([]map[string]interface{}, error) {
	return Search(ctx, "core.group", nil)
}
