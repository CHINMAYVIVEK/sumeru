// Package workflow documents the engine’s state-machine surface.
// Transition checks live in sumeru/core/orm (CanWorkflowTransition) to avoid import cycles.
// Metadata models: sys.workflow.transition (automation module).
package workflow
