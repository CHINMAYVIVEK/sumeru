package orm

import (
	"context"
	"encoding/json"
	"time"
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
		{Name: "published_at", Type: DateTime},
	}
}

func init() {
	RegisterModelWithModule(SysOutboxEvent{}, "base")
}

// EnqueueOutbox inserts an outbox row (best-effort; never fails the caller).
// When SecurityBypass is set, still records module-origin events if model exists.
func EnqueueOutbox(ctx context.Context, name string, actor int, payload map[string]interface{}) {
	if DB == nil || name == "" {
		return
	}
	inst, ok := Registry["sys.outbox.event"]
	if !ok || inst == nil {
		return
	}
	pj := ""
	if payload != nil {
		if b, err := json.Marshal(payload); err == nil {
			pj = string(b)
		}
	}
	vals := map[string]interface{}{
		"name":         name,
		"payload_json": pj,
		"actor":        actor,
		"created_at":   time.Now().UTC().Format(time.RFC3339),
	}
	bypass := ContextWithBypass(ctx, true)
	_, _ = Create(bypass, inst, vals)
}
