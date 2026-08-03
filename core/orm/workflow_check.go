package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// CanWorkflowTransition reports whether uid may move record id on model from→to.
// Prefers sys.workflow.transition; falls back to sys.approval_rule.
func CanWorkflowTransition(ctx context.Context, model string, recordID int, fromState, toState string, uid int) error {
	model = strings.TrimSpace(model)
	toState = strings.TrimSpace(toState)
	if model == "" || toState == "" {
		return fmt.Errorf("model and to_state required")
	}
	if SecurityBypass(ctx) || uid == superuserUID {
		return nil
	}
	if err := checkWorkflowTransitionRows(ctx, model, fromState, toState, uid); err != errNoWorkflowRows {
		return err
	}
	if recordID > 0 {
		return CheckStageApproval(ctx, model, recordID, toState)
	}
	return nil
}

var errNoWorkflowRows = fmt.Errorf("no workflow rows")

func checkWorkflowTransitionRows(ctx context.Context, model, fromState, toState string, uid int) error {
	if _, ok := Registry["sys.workflow.transition"]; !ok || DB == nil {
		return errNoWorkflowRows
	}
	tbl := GetTableName("sys.workflow.transition")
	rows, err := DB.QueryContext(ctx,
		`SELECT group_id, COALESCE(from_state,'') FROM `+tbl+
			` WHERE model = $1 AND to_state = $2 AND active = true`, model, toState)
	if err != nil {
		return errNoWorkflowRows
	}
	defer rows.Close()
	groups, err := EffectiveGroupIDs(ctx, uid)
	if err != nil {
		return err
	}
	n := 0
	matchedFrom := false
	allowed := false
	for rows.Next() {
		n++
		var gid sql.NullInt64
		var from string
		if err := rows.Scan(&gid, &from); err != nil {
			return err
		}
		if from != "" && from != fromState {
			continue
		}
		matchedFrom = true
		if !gid.Valid || gid.Int64 == 0 {
			allowed = true
			break
		}
		if intSliceContains(groups, int(gid.Int64)) {
			allowed = true
			break
		}
	}
	if n == 0 {
		return errNoWorkflowRows
	}
	if !matchedFrom || !allowed {
		return fmt.Errorf("workflow: transition %q → %q denied", fromState, toState)
	}
	return nil
}
