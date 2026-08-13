package web

import (
	"context"
	"encoding/json"
	"net/http"

	"sumeru/core/orm"
)

func writeJSON(w http.ResponseWriter, ctx context.Context, route string, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		WebLogEvent(ctx, route, "Failed to encode JSON response", "write", "partial", err, nil)
	}
}

func requireSystemAdmin(w http.ResponseWriter, r *http.Request, redirectOnDeny bool) bool {
	ctx := r.Context()
	uid := orm.SecurityUID(ctx)
	if uid <= 0 {
		uid = SessionUserID(r)
	}
	if orm.UserHasGroupXML(ctx, uid, "base.group_system") {
		return true
	}
	if redirectOnDeny {
		http.Redirect(w, r, "/web/home", http.StatusFound)
	} else {
		http.Error(w, "Forbidden", http.StatusForbidden)
	}
	return false
}

func requireModelAccess(w http.ResponseWriter, r *http.Request, model, perm string) bool {
	if err := orm.CheckModelAccess(r.Context(), orm.SecurityUID(r.Context()), model, perm); err != nil {
		WebLogEvent(r.Context(), r.URL.Path, "Model access denied", "access", "failure", err,
			map[string]interface{}{"resource": model, "permission": perm})
		http.Error(w, "Forbidden", http.StatusForbidden)
		return false
	}
	return true
}

func requireMenuAccess(w http.ResponseWriter, r *http.Request, menuXMLID string) bool {
	ctx := r.Context()
	uid := orm.SecurityUID(ctx)
	if uid <= 0 {
		uid = SessionUserID(r)
	}
	mid, _, err := orm.ResolveXmlId(ctx, menuXMLID)
	if err == nil && mid > 0 {
		if rec, err := orm.SearchOne(ctx, "sys.menu", map[string]interface{}{"id": mid}); err == nil {
			if orm.UserMayAccessMenu(ctx, uid, orm.AsString(rec["access_groups"])) {
				return true
			}
		}
	}
	if orm.UserHasGroupXML(ctx, uid, "base.group_system") {
		return true
	}
	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}
