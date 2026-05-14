package main

import (
	"fmt"
	"os"
	"strings"
)

// typeName converts technical_name to TechnicalName.
func typeName(technical string) string {
	parts := strings.Split(technical, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// humanTitle converts technical_name to Technical Name.
func humanTitle(technical string) string {
	parts := strings.Split(technical, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		low := strings.ToLower(p)
		parts[i] = strings.ToUpper(low[:1]) + low[1:]
	}
	return strings.Join(parts, " ")
}

// writeOrDie writes content to a file or exits.
func writeOrDie(path, content string) {
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", path, err)
		os.Exit(1)
	}
}
