package scheduler

import (
	"context"
	"strings"
	"sync"
)

// CronHandler runs when a sys.cron row fires and its code matches a registered handler.
type CronHandler func(ctx context.Context, payload map[string]interface{}) error

var (
	cronMu       sync.RWMutex
	cronHandlers = map[string]CronHandler{}
)

// RegisterCronHandler binds code (sys.cron code field) to fn.
func RegisterCronHandler(code string, fn CronHandler) {
	code = strings.TrimSpace(code)
	if code == "" || fn == nil {
		return
	}
	cronMu.Lock()
	defer cronMu.Unlock()
	cronHandlers[code] = fn
}

func lookupCronHandler(code string) CronHandler {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	cronMu.RLock()
	defer cronMu.RUnlock()
	return cronHandlers[code]
}

// ClearCronHandlers removes all handlers (tests).
func ClearCronHandlers() {
	cronMu.Lock()
	defer cronMu.Unlock()
	cronHandlers = map[string]CronHandler{}
}
