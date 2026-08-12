package applog

import (
	"strings"
	"time"
)

func parseLogTimezone(s string) (*time.Location, string) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "local") {
		return time.Local, "Local"
	}
	if strings.EqualFold(s, "utc") {
		return time.UTC, "UTC"
	}
	loc, err := time.LoadLocation(s)
	if err != nil {
		return time.UTC, "UTC"
	}
	return loc, s
}

func effectiveLocation() *time.Location {
	if logLocation != nil {
		return logLocation
	}
	return time.Local
}
