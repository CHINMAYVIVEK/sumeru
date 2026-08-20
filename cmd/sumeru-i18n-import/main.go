// Command sumeru-i18n-import loads translation rows from CSV into sys.translation.
package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"sumeru/core/orm"
	"sumeru/core/server/cliboot"
)

func main() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	inPath := flag.String("i", "translations.csv", "Input CSV path")
	flag.Parse()

	if err := cliboot.LoadConfig(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	f, err := os.Open(*inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "csv: %v\n", err)
		os.Exit(1)
	}
	if len(records) < 2 {
		fmt.Println("No data rows.")
		return
	}
	header := records[0]
	col := map[string]int{}
	for i, h := range header {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, req := range []string{"lang", "src", "value", "module"} {
		if _, ok := col[req]; !ok {
			fmt.Fprintf(os.Stderr, "missing column %q\n", req)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
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
	imported := 0
	for _, row := range records[1:] {
		if len(row) < len(header) {
			continue
		}
		lang := strings.TrimSpace(row[col["lang"]])
		src := strings.TrimSpace(row[col["src"]])
		val := row[col["value"]]
		mod := strings.TrimSpace(row[col["module"]])
		if lang == "" || src == "" {
			continue
		}
		_, err := db.ExecContext(ctx, `
			INSERT INTO "`+tableName+`" (lang, src, value, module)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (lang, src, module) DO UPDATE SET value = EXCLUDED.value`, lang, src, val, mod)
		if err != nil {
			fmt.Fprintf(os.Stderr, "insert: %v\n", err)
			os.Exit(1)
		}
		imported++
	}
	fmt.Printf("Imported %d translation rows from %s\n", imported, *inPath)
}
