package applog

import (
	"fmt"
	"os"
)

// BootstrapFatal writes a plain-text message to stderr and exits before SetupFromConfig.
func BootstrapFatal(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
