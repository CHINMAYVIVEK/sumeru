package web

import "testing"

func TestNormalizeGridListLayout(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "grid"},
		{"grid", "grid"},
		{"list", "list"},
		{"kanban", "grid"},
		{"GRID", "grid"},
		{" List ", "list"},
		{"table", "grid"},
	}
	for _, tc := range tests {
		if got := normalizeGridListLayout(tc.in); got != tc.want {
			t.Errorf("normalizeGridListLayout(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestModuleDisplayName(t *testing.T) {
	if got := moduleDisplayName("sale", ""); got != "sale" {
		t.Fatalf("empty display: got %q", got)
	}
	if got := moduleDisplayName("sale", "Sales"); got != "Sales" {
		t.Fatalf("with display: got %q", got)
	}
	if got := moduleDisplayName("sale", "  "); got != "sale" {
		t.Fatalf("whitespace display: got %q", got)
	}
}

func TestParseModuleRow(t *testing.T) {
	row, ok := parseModuleRow(map[string]interface{}{
		"name":         "crm",
		"display_name": "CRM",
		"state":        "installed",
		"application":  true,
		"active":       true,
	})
	if !ok {
		t.Fatal("expected ok")
	}
	if row.Name != "crm" || row.DisplayName != "CRM" || !row.Application || !row.Active {
		t.Fatalf("unexpected row: %+v", row)
	}
	_, ok = parseModuleRow(map[string]interface{}{"display_name": "No name"})
	if ok {
		t.Fatal("expected false for missing name")
	}
}
