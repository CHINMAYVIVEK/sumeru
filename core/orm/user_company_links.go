package orm

import (
	"context"
	"fmt"
)

// UserCompanyIDsForUser returns allowed company ids from core.user.company.rel.
func UserCompanyIDsForUser(ctx context.Context, userID int) ([]int, error) {
	if userID <= 0 || DB == nil {
		return nil, nil
	}
	tbl := GetTableName("core.user.company.rel")
	rows, err := DB.QueryContext(ctx, `SELECT company_id FROM `+tbl+` WHERE user_id = $1 ORDER BY company_id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var cid int
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		if cid > 0 {
			out = append(out, cid)
		}
	}
	return out, rows.Err()
}

// SetUserCompanyLinks replaces allowed-company membership for a user.
func SetUserCompanyLinks(ctx context.Context, userID int, companyIDs []int) error {
	if DB == nil {
		return fmt.Errorf("no database")
	}
	if userID <= 0 {
		return fmt.Errorf("invalid user id")
	}
	tbl := GetTableName("core.user.company.rel")
	if _, err := DB.ExecContext(ctx, `DELETE FROM `+tbl+` WHERE user_id = $1`, userID); err != nil {
		return err
	}
	seen := map[int]struct{}{}
	for _, cid := range companyIDs {
		if cid <= 0 {
			continue
		}
		if _, ok := seen[cid]; ok {
			continue
		}
		seen[cid] = struct{}{}
		if _, err := DB.ExecContext(ctx,
			`INSERT INTO `+tbl+` (user_id, company_id) VALUES ($1, $2) ON CONFLICT (user_id, company_id) DO NOTHING`,
			userID, cid); err != nil {
			return err
		}
	}
	return nil
}
