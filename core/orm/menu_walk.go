package orm

import (
	"context"
)

// MenuHasAncestor reports whether startID is rootID or a descendant (walk parent_id, max 64).
func MenuHasAncestor(ctx context.Context, startID, rootID int) bool {
	if startID <= 0 || rootID <= 0 || DB == nil {
		return false
	}
	cur := startID
	for i := 0; i < 64; i++ {
		if cur == rootID {
			return true
		}
		row, err := SearchOne(ctx, "sys.menu", map[string]interface{}{"id": cur})
		if err != nil {
			return false
		}
		pid, ok := CoerceInt64(row["parent_id"])
		if !ok || pid == 0 {
			return false
		}
		cur = int(pid)
	}
	return false
}
