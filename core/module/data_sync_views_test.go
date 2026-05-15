package module

import "testing"

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
		if got := inferSysViewTypeFromArch(tt.arch); got != tt.want {
			t.Errorf("inferSysViewTypeFromArch(%q) = %q; want %q", tt.arch, got, tt.want)
		}
	}
}
