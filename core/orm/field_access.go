package orm

import (
	"context"
	"database/sql"
	"fmt"
)

// FilterWritableFields removes keys the uid may not write (sys.field_access).
// If no rules exist for a field, write is allowed (opt-in deny).
func FilterWritableFields(ctx context.Context, uid int, model string, values map[string]interface{}) (map[string]interface{}, error) {
	if SecurityBypass(ctx) || uid == superuserUID || len(values) == 0 {
		return values, nil
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "write")
	if err != nil || len(denied) == 0 {
		return values, err
	}
	out := make(map[string]interface{}, len(values))
	for k, v := range values {
		if denied[k] {
			continue
		}
		out[k] = v
	}
	return out, nil
}

// CheckFieldWriteAccess errors if any key in values is write-denied.
func CheckFieldWriteAccess(ctx context.Context, uid int, model string, values map[string]interface{}) error {
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "write")
	if err != nil {
		return err
	}
	for k := range values {
		if denied[k] {
			return fmt.Errorf("field access denied: %s.%s", model, k)
		}
	}
	return nil
}

func fieldAccessDenied(ctx context.Context, uid int, model, op string) (map[string]bool, error) {
	out := map[string]bool{}
	if _, ok := Registry["sys.field_access"]; !ok || DB == nil {
		return out, nil
	}
	col := "perm_write"
	if op == "read" {
		col = "perm_read"
	}
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return nil, err
	}
	tbl := GetTableName("sys.field_access")
	rows, err := DB.QueryContext(ctx,
		`SELECT field_name, group_id, `+col+` FROM `+tbl+` WHERE model = $1`, model)
	if err != nil {
		return out, nil
	}
	defer rows.Close()
	// Collect rules per field; deny if user matches a rule with perm=false,
	// or if rules exist and none grant the permission.
	type rule struct {
		gid  int
		perm bool
	}
	byField := map[string][]rule{}
	for rows.Next() {
		var field string
		var gid sql.NullInt64
		var perm bool
		if err := rows.Scan(&field, &gid, &perm); err != nil {
			return nil, err
		}
		g := 0
		if gid.Valid {
			g = int(gid.Int64)
		}
		byField[field] = append(byField[field], rule{gid: g, perm: perm})
	}
	for field, rules := range byField {
		allowed := false
		matched := false
		for _, r := range rules {
			if r.gid == 0 || intSliceContains(groups, r.gid) {
				matched = true
				if r.perm {
					allowed = true
					break
				}
			}
		}
		if matched && !allowed {
			out[field] = true
		}
	}
	return out, rows.Err()
}

// StripDeniedReadFields removes read-denied fields from a record map.
func StripDeniedReadFields(ctx context.Context, uid int, model string, rec map[string]interface{}) map[string]interface{} {
	if SecurityBypass(ctx) || uid == superuserUID || rec == nil {
		return rec
	}
	denied, err := fieldAccessDenied(ctx, uid, model, "read")
	if err != nil || len(denied) == 0 {
		return rec
	}
	for k := range denied {
		delete(rec, k)
	}
	return rec
}
