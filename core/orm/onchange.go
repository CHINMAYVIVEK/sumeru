package orm

import (
	"context"
	"fmt"
	"strings"
)

// OnchangeResult is returned to the SWC client after a field onchange.
type OnchangeResult struct {
	Value   map[string]interface{} `json:"value"`
	Warning *OnchangeWarning       `json:"warning,omitempty"`
	Domain  map[string]interface{} `json:"domain,omitempty"`
}

// OnchangeWarning is an optional user-facing warning from an onchange handler.
type OnchangeWarning struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// OnchangeHandler runs when a form field changes in the client.
type OnchangeHandler func(ctx context.Context, values map[string]interface{}, field string) (OnchangeResult, error)

var onchangeRegistry = map[string]map[string]OnchangeHandler{}

// RegisterOnchange registers a field onchange handler for a model.
func RegisterOnchange(model, field string, fn OnchangeHandler) {
	model = strings.TrimSpace(model)
	field = strings.TrimSpace(field)
	if model == "" || field == "" || fn == nil {
		return
	}
	if onchangeRegistry[model] == nil {
		onchangeRegistry[model] = map[string]OnchangeHandler{}
	}
	onchangeRegistry[model][field] = fn
}

// RunOnchange executes a registered onchange handler.
func RunOnchange(ctx context.Context, model, field string, values map[string]interface{}) (OnchangeResult, error) {
	model = strings.TrimSpace(model)
	field = strings.TrimSpace(field)
	handlers, ok := onchangeRegistry[model]
	if !ok {
		return OnchangeResult{}, fmt.Errorf("no onchange handlers for model %q", model)
	}
	fn, ok := handlers[field]
	if !ok {
		return OnchangeResult{}, fmt.Errorf("no onchange handler for %s.%s", model, field)
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	return fn(ctx, values, field)
}

// HasOnchange reports whether a model/field pair has a registered handler.
func HasOnchange(model, field string) bool {
	handlers, ok := onchangeRegistry[strings.TrimSpace(model)]
	if !ok {
		return false
	}
	_, ok = handlers[strings.TrimSpace(field)]
	return ok
}
