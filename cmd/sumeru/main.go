//go:generate go run ../sumeru-import-gen -root ../.. -config sumeru.conf -out cmd/sumeru/zimports.go -package main

// Command sumeru is the default HTTP server entrypoint (library code lives under core/).
package main

import (
	"sumeru/core/server"
)

func main() {
	server.Run()
}
