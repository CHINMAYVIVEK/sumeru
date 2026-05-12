package base

import "sumeru/core/orm"

// RegisterModelInput registers a model with the global registry (table sync on startup).
type RegisterModelInput struct {
	Model Model
}

// RegisterModel wires RegisterModelInput to the ORM registry.
func RegisterModel(in RegisterModelInput) {
	if in.Model == nil {
		return
	}
	orm.RegisterModel(in.Model)
}
