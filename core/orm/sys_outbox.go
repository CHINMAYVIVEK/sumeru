package orm

import (
	"context"
	"encoding/json"
	"time"

	"sumeru/core/applog"
)

// SysOutboxEvent stores transactional outbox rows for reliable event publishing.
type SysOutboxEvent struct {
	ID          int    `orm:"id"`
	Name        string `orm:"name"`
	PayloadJSON string `orm:"payload_json"`
	Actor       int    `orm:"actor"`
	CreatedAt   string `orm:"created_at"`
	PublishedAt string `orm:"published_at"`
}

func (SysOutboxEvent) ModelName() string { return "sys.outbox.event" }
func (SysOutboxEvent) Fields() []FieldDefinition {
	return []FieldDefinition{
		{Name: "name", Type: Char, Required: true, Index: true},
		{Name: "payload_json", Type: Text},
		{Name: "actor", Type: Integer},
		{Name: "created_at", Type: DateTime, Required: true},
		// TODO: outbox drain worker + partial index WHERE published_at IS NULL
		{Name: "published_at", Type: DateTime, Index: true},
	}
}

func init() {
	RegisterModelWithModule(SysOutboxEvent{}, "base")
}

func outboxValues(name string, actor int, payload map[string]interface{}) map[string]interface{} {
	pj := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			pj = string(b)
		} else {
			applog.Warn(context.Background(), applog.Event{
				Message:   "Outbox payload marshal failed",
				Component: "orm",
				Operation: "outbox",
				Status:    "partial",
				Context:   map[string]interface{}{"event_name": name},
				Err:       err,
			})
		}
	}
	return map[string]interface{}{
		"name":         name,
		"payload_json": pj,
		"actor":        actor,
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	}
}

// EnqueueOutboxTx inserts on tx when non-nil.
func EnqueueOutboxTx(ctx context.Context, tx TxWrapper, name string, actor int, payload map[string]interface{}) {
	if name == "" {
		return
	}
	vals := outboxValues(name, actor, payload)
	_ = insertSideEffectRow(ctx, tx, "sys.outbox.event", vals)
}
