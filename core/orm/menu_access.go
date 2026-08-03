package orm

import (
	"context"
	"fmt"
)

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
