package sdk

import "sumeru/core/modelreg"

// MustRegister registers struct-based models for module. Called from generated code only.
func MustRegister(module string, models ...any) {
	modelreg.MustRegister(module, models...)
}
