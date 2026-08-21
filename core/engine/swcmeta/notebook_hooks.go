package swcmeta

import (
	"context"
	"strings"
)

// NotebookHook renders optional HTML for a notebook page title on a model.
type NotebookHook func(ctx context.Context, model string, record map[string]interface{}, readonly bool) string

var notebookHooks = map[string]map[string]NotebookHook{}

// RegisterNotebookHook registers SWC notebook page content for a model/page title.
func RegisterNotebookHook(model, pageTitle string, hook NotebookHook) {
	model = strings.TrimSpace(model)
	pageTitle = strings.ToLower(strings.TrimSpace(pageTitle))
	if model == "" || pageTitle == "" || hook == nil {
		return
	}
	if notebookHooks[model] == nil {
		notebookHooks[model] = map[string]NotebookHook{}
	}
	notebookHooks[model][pageTitle] = hook
}

// RenderNotebookHook returns hook HTML when registered.
func RenderNotebookHook(ctx context.Context, model, pageTitle string, record map[string]interface{}, readonly bool) string {
	model = strings.TrimSpace(model)
	pageTitle = strings.ToLower(strings.TrimSpace(pageTitle))
	hooks, ok := notebookHooks[model]
	if !ok {
		return ""
	}
	fn, ok := hooks[pageTitle]
	if !ok || fn == nil {
		return ""
	}
	return fn(ctx, model, record, readonly)
}
