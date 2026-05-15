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

// extractImpliedGroupXMLRefs returns XML id refs from a core.group implied_ids eval (Odoo-style link tuples).
func extractImpliedGroupXMLRefs(evalStr string) []string {
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
	refs := extractImpliedGroupXMLRefs(evalStr)
	if len(refs) == 0 {
		return fmt.Errorf("no (4, ref('…')) implied link commands in %q", evalStr)
	}
	implTbl := orm.GetTableName("core.group.implied")
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
