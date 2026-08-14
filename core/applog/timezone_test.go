package applog_test

import (
	"testing"
	"time"

	"sumeru/core/applog"
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
		loc, label := applog.ParseLogTimezone(tc.input)
		if loc.String() != tc.wantLoc.String() {
			t.Errorf("ParseLogTimezone(%q): got location %q, want %q", tc.input, loc, tc.wantLoc)
		}
		if label != tc.wantLabel {
			t.Errorf("ParseLogTimezone(%q): got label %q, want %q", tc.input, label, tc.wantLabel)
		}
	}
}

func TestEffectiveLocation_fallsBackToLocal(t *testing.T) {
	orig := applog.EffectiveLocationForTest()
	applog.SetLogLocationForTest(nil)
	defer applog.SetLogLocationForTest(orig)

	if got := applog.EffectiveLocationForTest(); got != time.Local {
		t.Errorf("EffectiveLocationForTest() with nil logLocation: got %v, want time.Local", got)
	}
}

func TestEffectiveLocation_returnsSet(t *testing.T) {
	orig := applog.EffectiveLocationForTest()
	applog.SetLogLocationForTest(time.UTC)
	defer applog.SetLogLocationForTest(orig)

	if got := applog.EffectiveLocationForTest(); got != time.UTC {
		t.Errorf("EffectiveLocationForTest() with UTC set: got %v, want UTC", got)
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
