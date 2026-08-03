package sdk

import (
	"strings"

	"sumeru/core/orm"
)

// RegisterModelInput registers a model with the global registry (table sync on startup).
type RegisterModelInput struct {
	Model  Model
	Module string // sys.module technical name of the declaring addon (required for addon models; use "base" for kernel types shipped with base)
}

// RegisterModel wires RegisterModelInput to the ORM registry.
func RegisterModel(in RegisterModelInput) {
	if in.Model == nil {
		return
	}
	mod := strings.TrimSpace(in.Module)
	orm.RegisterModelWithModule(in.Model, mod)
}
