package applog

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	uidMu       sync.RWMutex
	uidFromCtx  func(context.Context) int
	logLocation *time.Location // effective location for log_ts (set in SetupFromConfig)
	logEnabled  bool           // false → Nop logger, no structured output
	logTzName   string         // IANA or "UTC" / "Local" label for log_tz field
)

// RegisterUIDResolver wires user_id for L(ctx). Call once at process startup from server (e.g. orm.UIDFromContext).
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

// Enabled reports whether structured application logging is on (INI log_enabled).
func Enabled() bool {
	return logEnabled
}

func effectiveLocation() *time.Location {
	if logLocation != nil {
		return logLocation
	}
	return time.Local
}

// L returns a sugared logger with enforced fields: user_id, log_ts (RFC3339Nano in configured TZ), log_tz.
// When logging is disabled or Zap is not initialized, returns a no-op sugared logger.
func L(ctx context.Context) *zap.SugaredLogger {
	if !logEnabled {
		return zap.NewNop().Sugar()
	}
	s := Sugar()
	if s == nil {
		return zap.NewNop().Sugar()
	}
	loc := effectiveLocation()
	now := time.Now().In(loc)
	tzLabel := logTzName
	if tzLabel == "" {
		tzLabel = loc.String()
	}
	return s.With(
		"user_id", resolveUID(ctx),
		"log_ts", now.Format(time.RFC3339Nano),
		"log_tz", tzLabel,
	)
}

// ORMOp logs one ORM operation with enforced context fields (no-op when logging disabled).
func ORMOp(ctx context.Context, op, model string, err error, keysAndValues ...interface{}) {
	if !logEnabled {
		return
	}
	kvs := append([]interface{}{"op", op, "model", model}, keysAndValues...)
	if err != nil {
		kvs = append(kvs, "err", err)
		L(ctx).Errorw("orm", kvs...)
		return
	}
	L(ctx).Infow("orm", kvs...)
}
