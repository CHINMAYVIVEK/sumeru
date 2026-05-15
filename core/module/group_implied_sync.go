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

// syncCoreGroupImpliedFromEval parses implied_ids eval and inserts rows into core.group.implied.
func syncCoreGroupImpliedFromEval(ctx context.Context, moduleName string, groupID int, evalStr string) error {
	evalStr = strings.TrimSpace(evalStr)
	if groupID <= 0 || evalStr == "" {
		return nil
	}
	matches := impliedGroupLinkRefRE.FindAllStringSubmatch(evalStr, -1)
	if len(matches) == 0 {
		return fmt.Errorf("no (4, ref('…')) implied link commands in %q", evalStr)
	}
	implTbl := orm.GetTableName("core.group.implied")
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		ref := strings.TrimSpace(m[1])
		if ref == "" {
			continue
		}
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
