package orm

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"sumeru/core/applog"
	"sumeru/core/event"
)

var (
	outboxMu     sync.Mutex
	outboxCancel context.CancelFunc
)

// DrainOutboxOnce publishes pending outbox rows (up to 100) and marks them published.
func DrainOutboxOnce(ctx context.Context) int {
	if DB == nil {
		return 0
	}
	if _, ok := Registry["sys.outbox.event"]; !ok {
		return 0
	}
	bypass := ContextWithBypass(ctx, true)
	tbl := MustQuotedTableName("sys.outbox.event")
	rows, err := DB.QueryContext(bypass,
		`SELECT id, name, COALESCE(payload_json,''), COALESCE(actor,0) FROM `+tbl+
			` WHERE published_at IS NULL ORDER BY id LIMIT 100`)
	if err != nil {
		return 0
	}
	defer rows.Close()

	published := 0
	now := time.Now().UTC().Format(time.RFC3339)
	for rows.Next() {
		var id int64
		var name, payloadJSON string
		var actor int
		if err := rows.Scan(&id, &name, &payloadJSON, &actor); err != nil {
			continue
		}
		payload := map[string]interface{}{}
		if payloadJSON != "" {
			_ = json.Unmarshal([]byte(payloadJSON), &payload)
		}
		if errs := event.Publish(bypass, event.Event{Name: name, Actor: actor, Payload: payload}); len(errs) > 0 {
			applog.Warn(bypass, applog.Event{
				Message:   "outbox publish failed",
				Component: "orm",
				Operation: "outbox_drain",
				Status:    "failed",
				Context:   map[string]interface{}{"outbox_id": id, "event": name},
				Err:       errs[0],
			})
			continue
		}
		if _, err := DB.ExecContext(bypass,
			`UPDATE `+tbl+` SET published_at = $1 WHERE id = $2`, now, id); err != nil {
			continue
		}
		published++
	}
	return published
}

// StartOutboxDrain begins a background ticker that drains pending outbox rows.
func StartOutboxDrain(parent context.Context, every time.Duration) {
	outboxMu.Lock()
	defer outboxMu.Unlock()
	if outboxCancel != nil {
		return
	}
	if every <= 0 {
		every = 5 * time.Second
	}
	ctx, cancel := context.WithCancel(parent)
	outboxCancel = cancel
	go func() {
		t := time.NewTicker(every)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				DrainOutboxOnce(ctx)
			}
		}
	}()
}

// StopOutboxDrain halts the outbox drain goroutine.
func StopOutboxDrain() {
	outboxMu.Lock()
	defer outboxMu.Unlock()
	if outboxCancel != nil {
		outboxCancel()
		outboxCancel = nil
	}
}
