package web

import (
	"fmt"
	"log"
	"strings"
)

// WebLogf writes one standard-format line for web handlers (grep-friendly: component=web route=…).
func WebLogf(route, format string, args ...interface{}) {
	route = strings.TrimSpace(route)
	if route == "" {
		route = "-"
	}
	msg := fmt.Sprintf(format, args...)
	log.Printf("component=web route=%s %s", route, msg)
}
