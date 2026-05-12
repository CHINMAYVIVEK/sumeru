package server

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/config"
)

// Run parses flags, loads configuration, initializes persistence and modules,
// registers HTTP routes, and blocks in ListenAndServe.
//
// CLI flags (see README): -c, -d/--database, -i, -u, --http-port/-p, --stop-after-init.
func Run() {
	configPath := flag.String("c", "sumeru.conf", "Path to config file (INI)")
	installMods := flag.String("i", "", "Install modules (comma-separated). One or many: -i sales  OR  -i sales,crm")
	updateMods := flag.String("u", "", "Update modules (comma-separated). Use -u all for every installed module, or -u sales  OR  -u sales,crm")
	dbName := flag.String("d", "", "Database name; overrides db_name in config")
	dbNameLong := flag.String("database", "", "Database name (long form); same as -d if set")
	httpPort := flag.String("http-port", "", "HTTP listen port; overrides http_port in config")
	httpPortShort := flag.String("p", "", "HTTP port shorthand (same as --http-port)")
	stopAfterInit := flag.Bool("stop-after-init", false, "After -i / -u, exit without starting HTTP")
	flag.Parse()

	if err := LoadConfig(*configPath); err != nil {
		log.Fatalf("Critical Error: Failed to load configuration: %v", err)
	}
	if err := AbsPaths(); err != nil {
		log.Fatalf("Resolve paths: %v", err)
	}

	if err := setupLogFile(); err != nil {
		log.Fatalf("Log file: %v", err)
	}

	if s := strings.TrimSpace(*dbNameLong); s != "" {
		config.AppConfig.DbName = s
	} else if s := strings.TrimSpace(*dbName); s != "" {
		config.AppConfig.DbName = s
	}
	if s := strings.TrimSpace(*httpPortShort); s != "" {
		config.AppConfig.HttpPort = s
	} else if s := strings.TrimSpace(*httpPort); s != "" {
		config.AppConfig.HttpPort = s
	}
	if strings.TrimSpace(*dbName) != "" || strings.TrimSpace(*dbNameLong) != "" {
		log.Printf("Using database (CLI -d/--database): %s", config.AppConfig.DbName)
	}
	if strings.TrimSpace(*httpPort) != "" || strings.TrimSpace(*httpPortShort) != "" {
		log.Printf("Using HTTP port (CLI --http-port/-p): %s", config.AppConfig.HttpPort)
	}

	log.Printf("Addon roots: %v", config.AppConfig.AddonPaths)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.AppConfig.DbHost, config.AppConfig.DbPort, config.AppConfig.DbUser,
		config.AppConfig.DbPass, config.AppConfig.DbName, config.AppConfig.DbSslMode)
	InitDB(dsn)

	if err := SyncModels(); err != nil {
		log.Fatalf("Error syncing models: %v", err)
	}

	if err := orm.EnsureIrUiMenuModuleColumn(); err != nil {
		log.Fatalf("Schema migrate (ir_ui_menu.module): %v", err)
	}
	if err := orm.BackfillIrUiMenuModule(); err != nil {
		log.Printf("Warning: backfill ir_ui_menu.module: %v", err)
	}
	if err := orm.EnsureIrUiViewArchText(); err != nil {
		log.Printf("Note: ir_ui_view.arch column: %v", err)
	}
	if err := orm.EnsureResCompanyEnterpriseColumns(); err != nil {
		log.Fatalf("Schema migrate (res.company): %v", err)
	}
	if err := orm.EnsureResUsersEnterpriseColumns(); err != nil {
		log.Fatalf("Schema migrate (res.users): %v", err)
	}

	if err := LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		log.Fatalf("Addon load / convention: %v", err)
	}

	if err := RunModuleCLI(*installMods, *updateMods); err != nil {
		log.Fatalf("Module CLI (-i / -u): %v", err)
	}

	hadModuleOps := strings.TrimSpace(*installMods) != "" || strings.TrimSpace(*updateMods) != ""
	if *stopAfterInit && hadModuleOps {
		log.Println("stop-after-init: module operations finished, exiting.")
		os.Exit(0)
	}

	registerBrandingAndStatic()
	registerAppRoutes()

	log.Printf("Server starting on :%s...", config.AppConfig.HttpPort)
	if err := http.ListenAndServe(":"+config.AppConfig.HttpPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// setupLogFile opens AppConfig.LogFile for append (creating parent dirs). If set, log lines go to stderr and the file.
func setupLogFile() error {
	p := strings.TrimSpace(config.AppConfig.LogFile)
	if p == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return fmt.Errorf("mkdir log dir: %w", err)
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", p, err)
	}
	log.SetOutput(io.MultiWriter(os.Stderr, f))
	log.Printf("logging to %s", p)
	return nil
}
