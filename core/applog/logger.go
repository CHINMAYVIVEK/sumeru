package applog

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

var (
	uidMu       sync.RWMutex
	uidFromCtx  func(context.Context) int
	logLocation *time.Location
	logEnabled  bool
	logTzName   string
)

// RegisterUIDResolver wires user_id resolution for event context maps.
func RegisterUIDResolver(fn func(context.Context) int) {
	uidMu.Lock()
	defer uidMu.Unlock()
	uidFromCtx = fn
}

func resolveUID(ctx context.Context) int {
	uidMu.RLock()
	fn := uidFromCtx
	uidMu.RUnlock()
	if fn == nil || ctx == nil {
		return 0
	}
	return fn(ctx)
}

func baseLogger() *slog.Logger {
	if root != nil {
		return root
	}
	return slog.Default()
}

// LoggerFromContext returns the base logger. Per-event fields use the Event API;
// slog JSON handler supplies the canonical top-level "time" field (no log_ts duplication).
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if !logEnabled {
		return slog.New(slog.DiscardHandler)
	}
	return baseLogger()
}

// L is an alias for LoggerFromContext.
func L(ctx context.Context) *slog.Logger {
	return LoggerFromContext(ctx)
}

// LogORMOperation logs one ORM operation using the structured Event contract.
func LogORMOperation(ctx context.Context, op, model string, err error, keysAndValues ...interface{}) {
	if !logEnabled {
		return
	}
	ctxMap := map[string]interface{}{"resource": model}
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		if k, ok := keysAndValues[i].(string); ok {
			ctxMap[k] = keysAndValues[i+1]
		}
	}
	ev := Event{
		Component: "orm",
		Operation: op,
		Context:   ctxMap,
		Err:       err,
	}
	if err != nil {
		ev.Message = op + " on " + model + " failed"
		ev.Status = "failure"
		Error(ctx, ev)
		return
	}
	ev.Message = op + " on " + model + " completed"
	ev.Status = "success"
	Info(ctx, ev)
}

// Fatal logs at error level and exits the process.
func Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	attrs := keysAndValues
	if len(attrs) == 0 {
		Error(ctx, Event{Message: msg, Component: "server", Status: "failure"})
	} else {
		Error(ctx, Event{Message: msg, Component: "server", Status: "failure", Context: kvPairsToMap(attrs)})
	}
	os.Exit(1)
}

func kvPairsToMap(pairs []interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for i := 0; i+1 < len(pairs); i += 2 {
		if k, ok := pairs[i].(string); ok {
			out[k] = pairs[i+1]
		}
	}
	return out
}
