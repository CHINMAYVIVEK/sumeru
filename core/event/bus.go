// Package event provides a lightweight in-process event bus for modules.
package event

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	EventID   string
	Name      string
	Timestamp time.Time
	Actor     int
	Payload   map[string]interface{}
}

// Handler processes an event. Errors are logged by Publish callers.
type Handler func(ctx context.Context, ev Event) error

var (
	mu       sync.RWMutex
	handlers = map[string][]Handler{}
)

// Subscribe registers a handler for event name (e.g. "record.created").
func Subscribe(name string, h Handler) {
	if name == "" || h == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	handlers[name] = append(handlers[name], h)
}

// Publish delivers ev synchronously to all handlers for ev.Name.
func Publish(ctx context.Context, ev Event) []error {
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	mu.RLock()
	list := append([]Handler(nil), handlers[ev.Name]...)
	mu.RUnlock()
	var errs []error
	for _, h := range list {
		if err := h(ctx, ev); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// Clear removes all handlers (tests).
func Clear() {
	mu.Lock()
	defer mu.Unlock()
	handlers = map[string][]Handler{}
}
