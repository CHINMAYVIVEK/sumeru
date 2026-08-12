package orm

import (
	"context"
	"fmt"
	"strings"
)

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
	relTbl := GetTableName(tableRuleGroupRel)
	q := `SELECT r.id, r.domain_force, r.active, r.` + col + ` FROM ` + ruleTbl + ` r WHERE r.model = $1 AND r.active = true AND r.` + col + ` = true`
	rows, err := DB.QueryContext(ctx, q, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var globalDomains [][][]interface{}
	var groupDomains [][][]interface{}
	groupAllowAll := false

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
		dc := DomainContext{UID: uid}
		if cids, err := UserCompanyIDs(ctx, uid); err == nil {
			dc.CompanyIDs = cids
			if len(cids) > 0 {
				dc.CompanyID = cids[0]
			}
		}
		dom = SubstituteDomainContext(dom, dc)

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
			if !match {
				continue
			}
			// Empty domain on a matching group rule = allow all records (Manager “all documents”).
			if len(dom) == 0 {
				groupAllowAll = true
				continue
			}
			groupDomains = append(groupDomains, dom)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// (Global1 AND Global2 ...) AND (GroupRule1 OR GroupRule2 ...); empty group domain ⇒ no group filter.
	out := append([][][]interface{}(nil), globalDomains...)
	if groupAllowAll {
		return out, nil
	}

	switch len(groupDomains) {
	case 0:
		// No group rules applicable — no additional restriction.
	case 1:
		out = append(out, groupDomains[0])
	default:
		// OR-merge via SQL-friendly union: buildSearchWhereClause ANDs triples, so we
		// expand each group domain separately by relying on MergeRuleDomainsIntoSearch
		// callers that AND parts — for multi-group OR we flatten into one domain using
		// only supported ops by picking the least restrictive path at check time.
		// Search path: OR is approximated by merging with "|" markers consumed below.
		merged := mergeGroupDomainsOR(groupDomains)
		out = append(out, merged)
	}

	return out, nil
}

// mergeGroupDomainsOR builds a domain that matches if any leaf group domain matches.
// buildSearchWhereClause only ANDs; for multi-rule OR we use a special marker domain
// that MergeRuleDomainsIntoSearch / searchSQL expand. For CheckRecordRules, RecordMatchesDomain
// is updated to treat "|" prefixes. Prefer single-rule cases; for N>1 use OR via SQL IN expansion
// of the first field when all domains are single equality on the same field — otherwise keep
// first-match OR evaluation in CheckRecordRules only and for Search emit OR SQL.
func mergeGroupDomainsOR(groupDomains [][][]interface{}) [][]interface{} {
	if len(groupDomains) == 1 {
		return groupDomains[0]
	}
	// Marker: [["__or__", "=", "1"], ...flattened triples per domain as separate ApplicableRuleDomains parts]
	// Instead return a synthetic domain checked specially — use empty to mean OR-all was intended
	// when groupAllowAll is false: fold into check-time OR by storing as multiple parts.
	// Practical approach for own|all: Manager hits allow-all; User has one domain. Multi-OR rare.
	// Fallback: concatenate with OR markers for RecordMatchesDomain.
	merged := [][]interface{}{}
	for i := 0; i < len(groupDomains)-1; i++ {
		merged = append(merged, []interface{}{"|"})
	}
	for _, gd := range groupDomains {
		merged = append(merged, gd...)
	}
	return merged
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
