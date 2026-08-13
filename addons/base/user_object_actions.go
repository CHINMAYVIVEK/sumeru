package base

import (
	"context"
	"fmt"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

func init() {
	orm.RegisterObjectAction("core.user", "action_archive", archiveCoreUser)
	orm.RegisterObjectAction("core.user", "action_reset_password", resetPasswordCoreUser)
}

func archiveCoreUser(ctx context.Context, model string, id int, _ map[string]string) (string, error) {
	_ = model
	if err := orm.CheckModelAccess(ctx, orm.SecurityUID(ctx), "core.user", "write"); err != nil {
		return "", err
	}
	if _, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": id}); err != nil {
		return "", fmt.Errorf("record not found")
	}
	if err := orm.UpdateRecordByID(ctx, "core.user", id, map[string]interface{}{"active": false}); err != nil {
		return "", err
	}
	return "", nil
}

func resetPasswordCoreUser(ctx context.Context, model string, id int, vals map[string]string) (string, error) {
	_ = model
	_ = vals
	if !orm.UserHasGroupXML(ctx, orm.SecurityUID(ctx), "base.group_system") {
		return "", fmt.Errorf("access denied")
	}
	rec, err := orm.SearchOne(ctx, "core.user", map[string]interface{}{"id": id})
	if err != nil {
		return "", fmt.Errorf("record not found")
	}
	login := orm.AsString(rec["login"])
	applog.InfoMsg(ctx, "base", "user_action", "Password reset requested (email not yet wired)",
		map[string]interface{}{"user_id": id, "login": login})
	return "", nil
}
