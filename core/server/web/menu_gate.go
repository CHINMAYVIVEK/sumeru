package web

import (
	"context"

	"sumeru/core/orm"
)

func menuAccessAllowed(ctx context.Context, menuQuery string) bool {
	menuID, ok := parseMenuIDString(menuQuery)
	if !ok {
		return true
	}
	return userMayAccessMenuByID(ctx, orm.SecurityUID(ctx), menuID)
}
