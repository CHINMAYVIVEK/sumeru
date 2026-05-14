package orm

import (
	"context"
)

type contextKey string

const (
	uidKey    contextKey = "uid"
	bypassKey contextKey = "bypass"
)

// ContextWithUID returns a new context with the given user ID.
func ContextWithUID(ctx context.Context, uid int) context.Context {
	return context.WithValue(ctx, uidKey, uid)
}

// UIDFromContext returns the user ID from the context, or 0 if not set.
func UIDFromContext(ctx context.Context) int {
	if ctx == nil {
		return 0
	}
	if uid, ok := ctx.Value(uidKey).(int); ok {
		return uid
	}
	return 0
}

// ContextWithBypass returns a new context with the security bypass flag set.
func ContextWithBypass(ctx context.Context, bypass bool) context.Context {
	return context.WithValue(ctx, bypassKey, bypass)
}

// BypassFromContext returns true if security checks should be bypassed for this context.
func BypassFromContext(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if bypass, ok := ctx.Value(bypassKey).(bool); ok {
		return bypass
	}
	return false
}

// SecurityUID is a compatibility helper that returns the UID from a context.
// In the future, callers should use UIDFromContext(ctx) directly.
func SecurityUID(ctx context.Context) int {
	return UIDFromContext(ctx)
}

// SecurityBypass is a compatibility helper that returns the bypass flag from a context.
func SecurityBypass(ctx context.Context) bool {
	return BypassFromContext(ctx)
}
