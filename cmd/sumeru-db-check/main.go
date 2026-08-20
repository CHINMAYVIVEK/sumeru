// Command sumeru-db-check validates sumeru.conf and PostgreSQL connectivity.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"sumeru/core/server/cliboot"
	"sumeru/core/server/config"
)

func main() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	flag.Parse()

	if err := cliboot.LoadConfig(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := cliboot.OpenDB(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	c := config.AppConfig
	fmt.Printf("OK: database %s@%s:%s/%s reachable\n", c.DbUser, c.DbHost, c.DbPort, c.DbName)
}
