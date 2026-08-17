// Command sumeru-migrate runs versioned SQL migrations from addon migrations/ folders.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"sumeru/core/module"
	"sumeru/core/server/cliboot"
)

func main() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	moduleName := flag.String("module", "all", "Module name or 'all'")
	flag.Parse()

	ctx, err := cliboot.Init(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		os.Exit(1)
	}
	if strings.EqualFold(strings.TrimSpace(*moduleName), "all") {
		if err := module.RunAllMigrations(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("All migrations applied.")
		return
	}
	if err := module.RunModuleMigrations(ctx, strings.TrimSpace(*moduleName)); err != nil {
		fmt.Fprintf(os.Stderr, "migrate: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Migrations applied for %s\n", *moduleName)
}
