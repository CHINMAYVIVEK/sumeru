package module

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"sumeru/core/orm"
)

// impliedGroupLinkRefRE matches (4, ref('module.xml_id')) link tuples in core.group implied_ids eval strings.
var impliedGroupLinkRefRE = regexp.MustCompile(`\(\s*4\s*,\s*ref\s*\(\s*['"]([^'"]+)['"]\s*\)\s*\)`)

// ExtractImpliedGroupXMLRefs returns XML id refs from a core.group implied_ids eval (Odoo-style link tuples).
func ExtractImpliedGroupXMLRefs(evalStr string) []string {
	evalStr = strings.TrimSpace(evalStr)
	if evalStr == "" {
		return nil
	}
	matches := impliedGroupLinkRefRE.FindAllStringSubmatch(evalStr, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := strings.TrimSpace(m[1])
		if ref != "" {
			out = append(out, ref)
		}
	}
	return out
}

// syncCoreGroupImpliedFromEval parses implied_ids eval and inserts rows into core.group.implied.
func syncCoreGroupImpliedFromEval(ctx context.Context, moduleName string, groupID int, evalStr string) error {
	if groupID <= 0 || strings.TrimSpace(evalStr) == "" {
		return nil
	}
	refs := ExtractImpliedGroupXMLRefs(evalStr)
	if len(refs) == 0 {
		return fmt.Errorf("no (4, ref('…')) implied link commands in %q", evalStr)
	}
	implTbl := orm.MustQuotedTableName("core.group.implied")
	for _, ref := range refs {
		impliedID, err := resolveXMLIDInModule(ctx, moduleName, ref)
		if err != nil {
			return fmt.Errorf("resolve implied ref %q: %w", ref, err)
		}
		if impliedID <= 0 {
			return fmt.Errorf("resolve implied ref %q: id is 0", ref)
		}
		if impliedID == groupID {
			continue
		}
		_, err = orm.DB.ExecContext(ctx,
			`INSERT INTO `+implTbl+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`,
			groupID, impliedID)
		if err != nil {
			return fmt.Errorf("implied edge %d -> %d: %w", groupID, impliedID, err)
		}
	}
	return nil
}

// syncSysRuleGroupsFromEval parses groups eval [(4, ref('…'))] into sys.rule.group.rel.
func syncSysRuleGroupsFromEval(ctx context.Context, moduleName string, ruleID int, evalStr string) error {
	if ruleID <= 0 || strings.TrimSpace(evalStr) == "" {
		return nil
	}
	refs := ExtractImpliedGroupXMLRefs(evalStr)
	if len(refs) == 0 {
		return fmt.Errorf("no (4, ref('…')) group link commands in %q", evalStr)
	}
	relTbl := orm.MustQuotedTableName("sys.rule.group.rel")
	if _, err := orm.DB.ExecContext(ctx, `DELETE FROM `+relTbl+` WHERE rule_id = $1`, ruleID); err != nil {
		return err
	}
	for _, ref := range refs {
		gid, err := resolveXMLIDInModule(ctx, moduleName, ref)
		if err != nil {
			return fmt.Errorf("resolve rule group ref %q: %w", ref, err)
		}
		if gid <= 0 {
			continue
		}
		_, err = orm.DB.ExecContext(ctx,
			`INSERT INTO `+relTbl+` (rule_id, group_id) VALUES ($1, $2) ON CONFLICT (rule_id, group_id) DO NOTHING`,
			ruleID, gid)
		if err != nil {
			return fmt.Errorf("rule %d group %d: %w", ruleID, gid, err)
		}
	}
	return nil
}

// EnsureSystemImpliesManagerGroup links base.group_system → module Manager groups
// so Settings admins inherit every installed app Manager (and thus User) ladder.
func EnsureSystemImpliesManagerGroup(ctx context.Context, moduleName, recordXMLID string, groupID int) error {
	if groupID <= 0 {
		return nil
	}
	idLower := strings.ToLower(strings.TrimSpace(recordXMLID))
	if !strings.Contains(idLower, "manager") {
		return nil
	}
	sysGID, _, err := orm.ResolveXmlId(ctx, "base.group_system")
	if err != nil || sysGID <= 0 {
		return err
	}
	if sysGID == groupID {
		return nil
	}
	implTbl := orm.MustQuotedTableName("core.group.implied")
	_, err = orm.DB.ExecContext(ctx,
		`INSERT INTO `+implTbl+` (group_id, implied_group_id) VALUES ($1, $2) ON CONFLICT (group_id, implied_group_id) DO NOTHING`,
		sysGID, groupID)
	if err != nil {
		return fmt.Errorf("system implies %s.%s: %w", moduleName, recordXMLID, err)
	}
	return nil
}
