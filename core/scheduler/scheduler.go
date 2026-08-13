// Package scheduler runs sys.cron jobs on an interval.
package scheduler

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/event"
	"sumeru/core/orm"
)

var (
	mu     sync.Mutex
	cancel context.CancelFunc
)

// Start begins a background ticker that evaluates due cron rows.
func Start(parent context.Context, every time.Duration) {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		return
	}
	if every <= 0 {
		every = time.Minute
	}
	ctx, c := context.WithCancel(parent)
	cancel = c
	go loop(ctx, every)
}

// Stop halts the background ticker.
func Stop() {
	mu.Lock()
	defer mu.Unlock()
	if cancel != nil {
		cancel()
		cancel = nil
	}
}

func loop(ctx context.Context, every time.Duration) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			runDue(ctx)
		}
	}
}

func runDue(ctx context.Context) {
	if orm.DB == nil {
		return
	}
	if _, ok := orm.Registry["sys.cron"]; !ok {
		return
	}
	bypass := orm.ContextWithBypass(ctx, true)
	tbl := orm.MustQuotedTableName("sys.cron")
	now := time.Now().UTC()
	rows, err := orm.DB.QueryContext(bypass,
		`SELECT id, name, COALESCE(event_name,''), COALESCE(code,'') FROM `+tbl+
			` WHERE active = true AND (next_call IS NULL OR next_call <= $1)`, now,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var name, eventName, code string
		if err := rows.Scan(&id, &name, &eventName, &code); err != nil {
			continue
		}
		executeCron(bypass, id, name, eventName, code)
		interval := cronInterval(bypass, id)
		next := now.Add(interval)
		_, _ = orm.DB.ExecContext(bypass,
			`UPDATE `+tbl+` SET next_call = $1, last_call = $2 WHERE id = $3`,
			next, now, id,
		)
	}
}

func cronInterval(ctx context.Context, id int64) time.Duration {
	var mins sql.NullInt64
	_ = orm.DB.QueryRowContext(ctx,
		`SELECT interval_number FROM `+orm.MustQuotedTableName("sys.cron")+` WHERE id = $1`, id,
	).Scan(&mins)
	n := int(mins.Int64)
	if n <= 0 {
		n = 60
	}
	return time.Duration(n) * time.Minute
}

func executeCron(ctx context.Context, id int64, name, eventName, code string) {
	applog.L(ctx).Info("scheduler.cron", "id", id, "name", name, "code", code)
	payload := map[string]interface{}{"cron_id": id, "cron_name": name, "code": code}
	_ = event.Publish(ctx, event.Event{Name: "cron.tick", Payload: payload})
	if eventName = strings.TrimSpace(eventName); eventName != "" {
		_ = event.Publish(ctx, event.Event{Name: eventName, Payload: payload})
	}
}
