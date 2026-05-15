package web

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/applog"
)

// WebLogf writes a structured web log line with enforced fields from applog.L(ctx) (user_id, log_ts, log_tz).
func WebLogf(ctx context.Context, route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	msg := fmt.Sprintf(format, args...)
	applog.L(ctx).Infow("web", "route", route, "msg", msg)
}
