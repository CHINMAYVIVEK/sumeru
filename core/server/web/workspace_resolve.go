package web

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"sumeru/core/orm"
)

// ResolveWindowActionID returns sys.action.window id from query action (numeric or xml id) and/or menu_id.
func ResolveWindowActionID(ctx context.Context, actionIDStr, menuIDStr string) int {
	var actionID int
	if actionIDStr != "" {
		if id, err := strconv.Atoi(actionIDStr); err == nil {
			actionID = id
		} else {
			if resID, _, err := orm.ResolveXmlId(ctx, actionIDStr); err == nil {
				actionID = resID
			}
		}
	}
	if actionID == 0 && menuIDStr != "" {
		if menuID, err := strconv.Atoi(menuIDStr); err == nil {
			actionID = actionIDFromMenu(ctx, menuID)
		}
	}
	return actionID
}

func actionIDFromMenu(ctx context.Context, menuID int) int {
	menuData, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": menuID})
	if err != nil {
		return 0
	}
	if aID64, ok := orm.CoerceInt64(menuData["action_id"]); ok && aID64 != 0 {
		return int(aID64)
	}
	return firstDescendantActionID(ctx, menuID)
}

// firstDescendantActionID returns the first non-zero action_id in a depth-first walk
// of children ordered by sequence, then id.
func firstDescendantActionID(ctx context.Context, parentID int) int {
	rows, err := orm.DB.QueryContext(ctx,
		`SELECT id, action_id FROM `+orm.GetTableName("sys.menu")+
			` WHERE parent_id = $1 ORDER BY sequence ASC, id ASC`,
		parentID,
	)
	if err != nil {
		return 0
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var aid sql.NullInt64
		if err := rows.Scan(&cid, &aid); err != nil {
			continue
		}
		if aid.Valid && aid.Int64 != 0 {
			return int(aid.Int64)
		}
		if sub := firstDescendantActionID(ctx, cid); sub != 0 {
			return sub
		}
	}
	return 0
}

func menuIDPointsToAppLogs(ctx context.Context, menuIDStr string) bool {
	menuIDStr = strings.TrimSpace(menuIDStr)
	if menuIDStr == "" {
		return false
	}
	want, _, err := orm.ResolveXmlId(ctx, "base.menu_app_logs")
	if err != nil || want == 0 {
		return false
	}
	got, err := strconv.Atoi(menuIDStr)
	if err != nil {
		return false
	}
	return got == want
}

// isHomeMenuTree reports whether menu_id is base.menu_home_root or a descendant.
func isHomeMenuTree(ctx context.Context, menuIDStr string) bool {
	menuIDStr = strings.TrimSpace(menuIDStr)
	if menuIDStr == "" {
		return true
	}
	rootID, _, err := orm.ResolveXmlId(ctx, "base.menu_home_root")
	if err != nil || rootID == 0 {
		return false
	}
	cur, err := strconv.Atoi(menuIDStr)
	if err != nil || cur <= 0 {
		return false
	}
	for i := 0; i < 64 && cur > 0; i++ {
		if cur == rootID {
			return true
		}
		row, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": cur})
		if err != nil {
			return false
		}
		pid, ok := orm.CoerceInt64(row["parent_id"])
		if !ok || pid == 0 {
			return false
		}
		cur = int(pid)
	}
	return false
}

// actionWindowTargetModel returns the ORM technical model for a sys.action.window row (core_model).
func actionWindowTargetModel(actionData map[string]interface{}) string {
	return strings.TrimSpace(orm.AsString(actionData["core_model"]))
}
