// Package cliboot initializes config, database, and addons for standalone CLI tools.
package cliboot

import (
	"context"
	"fmt"
	"strings"

	"sumeru/core/orm"
	"sumeru/core/server/config"
	"sumeru/core/server"
)

// Init loads INI config, connects to PostgreSQL, syncs models, and discovers addons.
// Requires an initialized database (not setup mode).
func Init(configPath string) (context.Context, error) {
	if err := server.LoadConfig(configPath); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if err := server.AbsPaths(); err != nil {
		return nil, fmt.Errorf("paths: %w", err)
	}
	c := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)
	server.InitDatabase(dsn)
	if !orm.IsInitialized() {
		return nil, fmt.Errorf("database %q is not initialized; complete /setup first", c.DbName)
	}
	if err := server.SyncModels(); err != nil {
		return nil, fmt.Errorf("sync models: %w", err)
	}
	if err := orm.SyncRegistrySchema(); err != nil {
		return nil, fmt.Errorf("schema sync: %w", err)
	}
	if err := server.LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		return nil, fmt.Errorf("addons: %w", err)
	}
	ctx := orm.ContextWithBypass(context.Background(), true)
	if err := orm.EnsureDefaultGroupsAndImplied(); err != nil {
		return nil, fmt.Errorf("security groups: %w", err)
	}
	return ctx, nil
}

// InitOptionalDB loads config and addons without requiring DB (for list/depends-tree on disk only).
func InitOptionalDB(configPath string, requireDB bool) (context.Context, error) {
	if err := server.LoadConfig(configPath); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if err := server.AbsPaths(); err != nil {
		return nil, fmt.Errorf("paths: %w", err)
	}
	if err := server.LoadAddonPaths(config.AppConfig.AddonPaths); err != nil {
		return nil, fmt.Errorf("addons: %w", err)
	}
	if !requireDB {
		return context.Background(), nil
	}
	return Init(configPath)
}

// ContextWithUID returns ctx with security uid set for ORM calls.
func ContextWithUID(ctx context.Context, uid int) context.Context {
	if uid <= 0 {
		uid = 1
	}
	return orm.ContextWithUID(ctx, uid)
}

// DSN returns the primary PostgreSQL connection string from loaded config.
func DSN() string {
	c := config.AppConfig
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DbHost, c.DbPort, c.DbUser, c.DbPass, c.DbName, c.DbSslMode)
}

// ReadReplicaDSN returns read replica DSN when configured, else primary DSN.
func ReadReplicaDSN() string {
	if s := strings.TrimSpace(config.AppConfig.DbReadReplicaDSN); s != "" {
		return s
	}
	return DSN()
}
