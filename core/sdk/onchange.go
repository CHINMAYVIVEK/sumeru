package sdk

import (
	"context"

	"sumeru/core/orm"
)

// RegisterOnchange wires a field onchange handler to the ORM registry.
func RegisterOnchange(model, field string, fn func(ctx context.Context, values map[string]interface{}, field string) (orm.OnchangeResult, error)) {
	orm.RegisterOnchange(model, field, fn)
}
