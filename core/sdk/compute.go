package sdk

import "context"

// ComputeContext is passed to compute handlers registered via orm.RegisterCompute.
type ComputeContext struct {
	Ctx context.Context
	ID  int
}

// NewComputeContext builds a compute context for a record.
func NewComputeContext(ctx context.Context, id int) ComputeContext {
	return ComputeContext{Ctx: ctx, ID: id}
}
