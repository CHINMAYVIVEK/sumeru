package runtime

import (
	"context"
	"sync"

	"sumeru/core/event"
	"sumeru/core/orm"
)

// Runtime holds process-scoped services. Package globals remain for bootstrap;
// new code should prefer injecting *Runtime instead of adding more package-level singletons.
type Runtime struct {
	DB       orm.DBWrapper
	Registry map[string]orm.Model
	Events   *EventBus
	mu       sync.RWMutex
	pinned   bool // SetDefault: do not overwrite Registry/DB from orm globals
}

// EventBus wraps the package event bus so Runtime can own the publish surface.
type EventBus struct{}

// Publish delivers to the process event bus.
func (EventBus) Publish(ctx context.Context, ev event.Event) []error {
	return event.Publish(ctx, ev)
}

var (
	defaultMu sync.RWMutex
	defaultRT *Runtime
)

// Default returns the process default Runtime (lazily bound to orm globals).
func Default() *Runtime {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultRT == nil {
		defaultRT = &Runtime{
			DB:       orm.DB,
			Registry: orm.Registry,
			Events:   &EventBus{},
		}
		return defaultRT
	}
	if !defaultRT.pinned {
		defaultRT.DB = orm.DB
		defaultRT.Registry = orm.Registry
	}
	return defaultRT
}

// SyncFromGlobals refreshes the default Runtime from orm package globals (after InitDB).
func SyncFromGlobals() {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if defaultRT == nil {
		defaultRT = &Runtime{Events: &EventBus{}}
	}
	if defaultRT.pinned {
		return
	}
	defaultRT.DB = orm.DB
	defaultRT.Registry = orm.Registry
}

// SetDefault replaces the process default (tests). Pass nil to clear.
// Non-nil runtimes are pinned so Default() will not overwrite Registry/DB from package globals.
func SetDefault(rt *Runtime) {
	defaultMu.Lock()
	if rt != nil {
		rt.pinned = true
	}
	defaultRT = rt
	defaultMu.Unlock()
}

// New constructs an isolated Runtime for tests.
func New(db orm.DBWrapper, registry map[string]orm.Model) *Runtime {
	if registry == nil {
		registry = map[string]orm.Model{}
	}
	return &Runtime{DB: db, Registry: registry, Events: &EventBus{}, pinned: true}
}

// Model looks up a model on this runtime's registry.
func (rt *Runtime) Model(name string) (orm.Model, bool) {
	if rt == nil {
		return nil, false
	}
	rt.mu.RLock()
	defer rt.mu.RUnlock()
	m, ok := rt.Registry[name]
	return m, ok
}

// RegisterModel stores a model on this runtime (tests / isolated loaders).
func (rt *Runtime) RegisterModel(m orm.Model) {
	if rt == nil || m == nil {
		return
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.Registry == nil {
		rt.Registry = map[string]orm.Model{}
	}
	rt.Registry[m.ModelName()] = m
}
