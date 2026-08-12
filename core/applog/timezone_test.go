package applog

import (
	"testing"
	"time"
)

func TestParseLogTimezone(t *testing.T) {
	cases := []struct {
		input     string
		wantLoc   *time.Location
		wantLabel string
	}{
		{"", time.Local, "Local"},
		{"local", time.Local, "Local"},
		{"LOCAL", time.Local, "Local"},
		{"utc", time.UTC, "UTC"},
		{"UTC", time.UTC, "UTC"},
		{"Asia/Kolkata", mustLoadLocation(t, "Asia/Kolkata"), "Asia/Kolkata"},
		{"invalid/zone", time.UTC, "UTC"},
	}

	for _, tc := range cases {
		loc, label := parseLogTimezone(tc.input)
		if loc.String() != tc.wantLoc.String() {
			t.Errorf("parseLogTimezone(%q): got location %q, want %q", tc.input, loc, tc.wantLoc)
		}
		if label != tc.wantLabel {
			t.Errorf("parseLogTimezone(%q): got label %q, want %q", tc.input, label, tc.wantLabel)
		}
	}
}

func TestEffectiveLocation_fallsBackToLocal(t *testing.T) {
	orig := logLocation
	logLocation = nil
	defer func() { logLocation = orig }()

	if got := effectiveLocation(); got != time.Local {
		t.Errorf("effectiveLocation() with nil logLocation: got %v, want time.Local", got)
	}
}

func TestEffectiveLocation_returnsSet(t *testing.T) {
	orig := logLocation
	logLocation = time.UTC
	defer func() { logLocation = orig }()

	if got := effectiveLocation(); got != time.UTC {
		t.Errorf("effectiveLocation() with UTC set: got %v, want UTC", got)
	}
}

func mustLoadLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Skipf("timezone %q not available on this system: %v", name, err)
	}
	return loc
}
