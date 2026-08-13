package module_test

import (
	"testing"

	"sumeru/core/module"
)

func TestInferSysViewTypeFromArch(t *testing.T) {
	tests := []struct {
		arch string
		want string
	}{
		{"<tree string=\"X\"><field name=\"a\"/></tree>", "tree"},
		{"  <form><sheet/></form>", "form"},
		{"<kanban><field name=\"n\"/></kanban>", "kanban"},
		{"<view type=\"tree\" model=\"m\"><tree/></view>", "tree"},
		{"", ""},
		{"<unknown/>", ""},
	}
	for _, tt := range tests {
		if got := module.InferSysViewTypeFromArch(tt.arch); got != tt.want {
			t.Errorf("InferSysViewTypeFromArch(%q) = %q; want %q", tt.arch, got, tt.want)
		}
	}
}
