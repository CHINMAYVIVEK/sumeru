package sdk

import "sumeru/core/orm"

// AsStringInput coerces a driver value to a printable string.
type AsStringInput struct {
	V interface{}
}

// AsString delegates to the ORM helper.
func AsString(in AsStringInput) string {
	return orm.AsString(in.V)
}

// CoerceInt64Input coerces a driver value to int64.
type CoerceInt64Input struct {
	V interface{}
}

// CoerceInt64 delegates to the ORM helper.
func CoerceInt64(in CoerceInt64Input) (int64, bool) {
	return orm.CoerceInt64(in.V)
}
