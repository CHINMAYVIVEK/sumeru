package render

import (
	"context"
	"strings"

	"sumeru/core/orm"
)

// RootMenuIDForModule returns the root sys.menu id for an installed module (parent_id IS NULL).
func RootMenuIDForModule(ctx context.Context, moduleName string) int {
	if orm.DB == nil || strings.TrimSpace(moduleName) == "" {
		return 0
	}
	table := orm.MustQuotedTableName("sys.menu")
	query := `SELECT id FROM ` + table + ` WHERE module = $1 AND parent_id IS NULL ORDER BY sequence ASC, id ASC LIMIT 1`
	var id int
	if err := orm.DB.QueryRowContext(ctx, query, strings.TrimSpace(moduleName)).Scan(&id); err != nil {
		return 0
	}
	return id
}

// IconLetterFromName returns the first letter of displayName for app launcher tiles.
func IconLetterFromName(displayName string) string {
	if runes := []rune(strings.TrimSpace(displayName)); len(runes) > 0 {
		return strings.ToUpper(string(runes[0]))
	}
	return "?"
}
