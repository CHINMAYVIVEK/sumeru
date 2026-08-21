// Command sumeru-i18n-export dumps sys.translation rows to CSV.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/cliboot"
)

func main() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	outPath := flag.String("o", "translations.csv", "Output CSV path")
	flag.Parse()

	if err := cliboot.LoadConfig(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := cliboot.OpenDB(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "db: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	tableName, err := orm.ModelToTableName("sys.translation")
	if err != nil {
		fmt.Fprintf(os.Stderr, "table name: %v\n", err)
		os.Exit(1)
	}

	rows, err := db.QueryContext(ctx,
		`SELECT lang, src, value, module FROM "`+tableName+`" ORDER BY lang, module, src`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(1)
	}
	defer rows.Close()

	f, err := os.Create(*outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create output: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"lang", "src", "value", "module"})
	count := 0
	for rows.Next() {
		var lang, src, value, module string
		if err := rows.Scan(&lang, &src, &value, &module); err != nil {
			fmt.Fprintf(os.Stderr, "scan: %v\n", err)
			os.Exit(1)
		}
		_ = w.Write([]string{lang, src, value, module})
		count++
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "write csv: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Exported %d translation rows to %s\n", count, *outPath)
}
