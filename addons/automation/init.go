package automation

import (
	"context"
	"log"

	_ "sumeru/addons/automation/models"
	"sumeru/core/event"
	"sumeru/core/orm"
)

func init() {
	log.Println("Sumeru Automation Addon Loaded")
	event.Subscribe("record.created", runServerActionsForEvent)
	event.Subscribe("record.updated", runServerActionsForEvent)
	event.Subscribe("cron.tick", runServerActionsForEvent)
}

// WIP: server-action code execution is not implemented; matching rows are logged only.
func runServerActionsForEvent(ctx context.Context, ev event.Event) error {
	if orm.DB == nil {
		return nil
	}
	if _, ok := orm.Registry["sys.server_action"]; !ok {
		return nil
	}
	bypass := orm.ContextWithBypass(ctx, true)
	rows, err := orm.Search(bypass, "sys.server_action", [][]interface{}{
		{"event_name", "=", ev.Name},
		{"active", "=", true},
	})
	if err != nil {
		return err
	}
	for _, row := range rows {
		log.Printf("[automation] server action %v on event %s", row["name"], ev.Name)
	}
	return nil
}
