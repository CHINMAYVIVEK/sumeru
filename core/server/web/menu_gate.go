package web

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

func menuAccessAllowed(ctx context.Context, menuIDStr string) bool {
	mid, err := strconv.Atoi(strings.TrimSpace(menuIDStr))
	if err != nil || mid <= 0 {
		return true
	}
	uid := orm.SecurityUID(ctx)
	rec, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": mid})
	if err != nil {
		return false
	}
	return orm.UserMayAccessMenu(ctx, uid, orm.AsString(rec["access_groups"]))
}
