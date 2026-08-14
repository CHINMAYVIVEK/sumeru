package web

import (
	"context"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

func menuAccessAllowed(ctx context.Context, menuIDStr string) bool {
	menuID, err := strconv.Atoi(strings.TrimSpace(menuIDStr))
	if err != nil || menuID <= 0 {
		return true
	}
	return userMayAccessMenuByID(ctx, orm.SecurityUID(ctx), menuID)
}
