package orm

import (
	"context"
	"database/sql"
)

// UserCompanyIDs returns allowed company ids for uid (M2M + current company_id).
func UserCompanyIDs(ctx context.Context, uid int) ([]int64, error) {
	if uid <= 0 || DB == nil {
		return nil, nil
	}
	seen := map[int64]struct{}{}
	var out []int64
	add := func(id int64) {
		if id <= 0 {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	var companyID sql.NullInt64
	_ = DB.QueryRowContext(ctx,
		`SELECT company_id FROM `+MustQuotedTableName("core.user")+` WHERE id = $1`, uid,
	).Scan(&companyID)
	if companyID.Valid {
		add(companyID.Int64)
	}
	rel := MustQuotedTableName("core.user.company.rel")
	rows, err := DB.QueryContext(ctx, `SELECT company_id FROM `+rel+` WHERE user_id = $1`, uid)
	if err != nil {
		// Join table may not exist yet during early bootstrap.
		return out, nil
	}
	defer rows.Close()
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return out, err
		}
		add(cid)
	}
	return out, rows.Err()
}

// UserAllowedCompany returns true if cid is in the user's company set (or user has none configured → allow current only).
func UserAllowedCompany(ctx context.Context, uid int, cid int64) bool {
	if uid == superuserUID {
		return true
	}
	ids, err := UserCompanyIDs(ctx, uid)
	if err != nil || len(ids) == 0 {
		return true
	}
	for _, id := range ids {
		if id == cid {
			return true
		}
	}
	return false
}

// ActiveCompanyIDForUser reads core.user.company_id directly (no ORM Search/log path).
func ActiveCompanyIDForUser(ctx context.Context, uid int) int64 {
	if uid <= 0 || DB == nil {
		return 0
	}
	var companyID sql.NullInt64
	_ = DB.QueryRowContext(ctx,
		`SELECT company_id FROM `+MustQuotedTableName("core.user")+` WHERE id = $1`, uid,
	).Scan(&companyID)
	if !companyID.Valid {
		return 0
	}
	return companyID.Int64
}
