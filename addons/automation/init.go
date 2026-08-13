package automation

import (
	"context"
	"fmt"

	_ "sumeru/addons/automation/models"
	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	event.Subscribe("record.created", runServerActionsForEvent)
	event.Subscribe("record.updated", runServerActionsForEvent)
	event.Subscribe("cron.tick", runServerActionsForEvent)
}

// TODO: run server actions (log matches for now)
func runServerActionsForEvent(ctx context.Context, ev event.Event) error {
	if orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["sys.server.action"]; !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "sys.server.action", [][]interface{}{
		{"event_name", "=", ev.Name},
		{"active", "=", true},
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		applog.DebugMsg(ctx, "module", "automation",
			fmt.Sprintf("server action %v on event %s", row["name"], ev.Name),
			map[string]interface{}{"module": "automation", "event": ev.Name})
	}
	return nil
}
