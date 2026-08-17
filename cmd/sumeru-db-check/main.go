// Command sumeru-db-check validates sumeru.conf and PostgreSQL connectivity.
package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"

	"sumeru/core/server/config"
)

func main() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	flag.Parse()

	if err := config.LoadConfig(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if err := config.AbsPaths(); err != nil {
		fmt.Fprintf(os.Stderr, "paths: %v\n", err)
		os.Exit(1)
	}

	c := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db open: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "db ping: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("OK: database %s@%s:%s/%s reachable\n", c.DbUser, c.DbHost, c.DbPort, c.DbName)
}
