package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"sumeru/core/applog"
	"sumeru/core/orm"
	"sumeru/core/scheduler"
	"sumeru/core/server/config"
	"sumeru/core/server/web"
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

	if err := applog.SetupFromConfig(&config.AppConfig); err != nil {
		log.Fatalf("Logging: %v", err)
	}
	defer applog.Sync()
	applog.RegisterUIDResolver(orm.UIDFromContext)

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
		applog.L(context.Background()).Infow("server", "msg", "database override", "db", config.AppConfig.DbName)
	}
	if strings.TrimSpace(*httpPort) != "" || strings.TrimSpace(*httpPortShort) != "" {
		applog.L(context.Background()).Infow("server", "msg", "http port override", "port", config.AppConfig.HttpPort)
	}

	applog.L(context.Background()).Infow("server", "msg", "addon roots", "paths", config.AppConfig.AddonPaths)

	databaseSource := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.AppConfig.DbHost, config.AppConfig.DbPort, config.AppConfig.DbUser,
		config.AppConfig.DbPass, config.AppConfig.DbName, config.AppConfig.DbSslMode)
	InitDB(databaseSource)

	if err := LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		log.Fatalf("Addon load / convention: %v", err)
	}

	if !orm.IsInitialized() {
		log.Println("Database is not initialized. Starting in SETUP MODE.")
		log.Println("Visit http://localhost:" + config.AppConfig.HttpPort + "/setup to initialize the system.")

		registerBrandingAndStatic()
		registerSetupRoutes()

		log.Printf("Server starting on :%s (SETUP MODE)...", config.AppConfig.HttpPort)
		// Use default mux for setup
		if err := http.ListenAndServe(":"+config.AppConfig.HttpPort, nil); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
		return
	}

	// NORMAL STARTUP
	if err := SyncModels(); err != nil {
		log.Fatalf("Error syncing models: %v", err)
	}

	if err := orm.SyncRegistrySchema(); err != nil {
		log.Fatalf("Schema sync (model-driven): %v", err)
	}

	if err := orm.BackfillSysMenuModule(); err != nil {
		log.Printf("Warning: backfill sys.menu.module: %v", err)
	}
	if err := orm.FixSysMenuSelfParent(); err != nil {
		log.Printf("Warning: fix sys.menu self-parent: %v", err)
	}
	if err := orm.EnsureSysViewArchText(); err != nil {
		log.Printf("Note: sys.view.arch column: %v", err)
	}
	if err := orm.EnsureMailMessageModelResIndex(); err != nil {
		log.Fatalf("Schema migrate (mail.message index): %v", err)
	}

	if err := orm.EnsureDefaultGroupsAndImplied(); err != nil {
		log.Fatalf("Default security groups: %v", err)
	}

	if err := RunModuleCLI(*installMods, *updateMods); err != nil {
		log.Fatalf("Module CLI (-i / -u): %v", err)
	}

	if err := orm.EnsureBootstrapSecurity(); err != nil {
		log.Fatalf("Security bootstrap: %v", err)
	}

	hadModuleOperations := strings.TrimSpace(*installMods) != "" || strings.TrimSpace(*updateMods) != ""
	if *stopAfterInit && hadModuleOperations {
		log.Println("stop-after-init: module operations finished, exiting.")
		os.Exit(0)
	}

	registerBrandingAndStatic()
	registerAppRoutes()
	scheduler.Start(context.Background(), time.Minute)

	log.Printf("Server starting on :%s...", config.AppConfig.HttpPort)
	appHandler := web.SecurityMiddleware(nil)
	if err := http.ListenAndServe(":"+config.AppConfig.HttpPort, appHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
