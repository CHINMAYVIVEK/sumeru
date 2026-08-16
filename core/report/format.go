package report

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"sumeru/core/orm"
)

type CellFormatter func(ctx context.Context, model, field string, raw interface{}) string

var (
	formatterMu sync.RWMutex
	formatters  = map[string]CellFormatter{}
)

// RegisterCellFormatter registers a per-model export cell formatter.
func RegisterCellFormatter(model string, fn CellFormatter) {
	if model == "" || fn == nil {
		return
	}
	formatterMu.Lock()
	formatters[model] = fn
	formatterMu.Unlock()
}

func formatCell(ctx context.Context, model, field string, raw interface{}) string {
	formatterMu.RLock()
	fn := formatters[model]
	formatterMu.RUnlock()
	if fn != nil {
		if s := fn(ctx, model, field, raw); s != "" {
			return s
		}
	}
	if raw == nil {
		return ""
	}
	switch v := raw.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		return fmt.Sprintf("%g", v)
	case int, int32, int64:
		return fmt.Sprintf("%d", v)
	default:
		return strings.TrimSpace(orm.AsString(raw))
	}
}
