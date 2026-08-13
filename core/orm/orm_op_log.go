package orm

import (
	"context"
	"time"

	"sumeru/core/applog"
	"sumeru/core/metrics"
)

// logORMOperation records Prometheus-style duration metrics and structured ORM logs.
func logORMOperation(ctx context.Context, start time.Time, op, model string, err error, keysAndValues ...interface{}) {
	d := time.Since(start)
	metrics.Inc("sumeru_orm_ops_total")
	metrics.ObserveDuration("sumeru_db_query_duration_seconds", d)
	kvs := append([]interface{}{"sql_duration_ms", d.Milliseconds()}, keysAndValues...)
	applog.LogORMOperation(ctx, op, model, err, kvs...)
}
