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
	logLocation *time.Location // effective location for log_ts (set in SetupFromConfig)
	logEnabled  bool           // false → discard
	logTzName   string         // IANA or "UTC" / "Local" label for log_tz field
)

// RegisterUIDResolver wires user_id for LoggerFromContext. Call once at process startup.
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

// LoggerFromContext returns a slog logger with user_id, request_id, log_ts, log_tz.
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if !logEnabled {
		return slog.New(slog.DiscardHandler)
	}
	loc := effectiveLocation()
	now := time.Now().In(loc)
	tzLabel := logTzName
	if tzLabel == "" {
		tzLabel = loc.String()
	}
	return baseLogger().With(
		"user_id", resolveUID(ctx),
		"request_id", RequestIDFromContext(ctx),
		"log_ts", now.Format(time.RFC3339Nano),
		"log_tz", tzLabel,
	)
}

// L is an alias for LoggerFromContext.
func L(ctx context.Context) *slog.Logger {
	return LoggerFromContext(ctx)
}

// LogORMOperation logs one ORM operation with enforced context fields.
func LogORMOperation(ctx context.Context, op, model string, err error, keysAndValues ...interface{}) {
	if !logEnabled {
		return
	}
	attrs := append([]interface{}{"op", op, "model", model}, keysAndValues...)
	if err != nil {
		attrs = append(attrs, "err", err)
		LoggerFromContext(ctx).Error("orm", attrs...)
		return
	}
	LoggerFromContext(ctx).Info("orm", attrs...)
}

// Fatal logs at error level and exits the process.
func Fatal(ctx context.Context, msg string, keysAndValues ...interface{}) {
	LoggerFromContext(ctx).Error(msg, keysAndValues...)
	os.Exit(1)
}
