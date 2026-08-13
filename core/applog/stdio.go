package applog

import "context"

// InfoMsg logs a structured info event with the logging contract.
func InfoMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Info(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "success",
		Context:   ctxFields,
	})
}

// WarnMsg logs a structured warning event with the logging contract.
func WarnMsg(ctx context.Context, component, operation, message string, err error, ctxFields map[string]interface{}) {
	Warn(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "partial",
		Context:   ctxFields,
		Err:       err,
	})
}

// DebugMsg logs a structured debug event with the logging contract.
func DebugMsg(ctx context.Context, component, operation, message string, ctxFields map[string]interface{}) {
	Debug(ctx, Event{
		Message:   message,
		Component: component,
		Operation: operation,
		Status:    "success",
		Context:   ctxFields,
	})
}
