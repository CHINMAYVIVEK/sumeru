package server

import (
	"log"
	"os"
	"syscall"
	"time"
)

// selfRestart re-execs the current binary with the original arguments and
// environment, replacing the current process (Unix only).
//
// TODO(windows): Windows requires a different approach (spawn child + kill self)
// since syscall.Exec is not available on Windows.
func selfRestart(delayMs int) {
	time.Sleep(time.Duration(delayMs) * time.Millisecond)
	log.Println("Server: performing self-restart after setup…")
	if err := syscall.Exec(os.Args[0], os.Args, os.Environ()); err != nil {
		log.Fatalf("Self-restart failed: %v", err)
	}
}
