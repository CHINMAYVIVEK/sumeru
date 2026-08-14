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
		{"<list string=\"X\"><field name=\"a\"/></list>", "list"},
		{"  <form><sheet/></form>", "form"},
		{"<kanban><field name=\"n\"/></kanban>", "kanban"},
		{"<view type=\"list\" model=\"m\"><list/></view>", "list"},
		{"", ""},
		{"<unknown/>", ""},
	}
	for _, tt := range tests {
		if got := module.InferSysViewTypeFromArch(tt.arch); got != tt.want {
			t.Errorf("InferSysViewTypeFromArch(%q) = %q; want %q", tt.arch, got, tt.want)
		}
	}
}
