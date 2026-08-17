package module

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"sumeru/core/orm"
)

const migrationTable = "sys_migration"

// EnsureMigrationTable creates the migration tracking table if missing.
func EnsureMigrationTable(ctx context.Context) error {
	if orm.DB == nil {
		return fmt.Errorf("database not connected")
	}
	_, err := orm.DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS sys_migration (
			module TEXT NOT NULL,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (module, name)
		)`)
	return err
}

func migrationApplied(ctx context.Context, moduleName, name string) (bool, error) {
	var n int
	err := orm.DB.QueryRowContext(ctx,
		`SELECT 1 FROM sys_migration WHERE module = $1 AND name = $2`, moduleName, name,
	).Scan(&n)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func markMigrationApplied(ctx context.Context, moduleName, name string) error {
	_, err := orm.DB.ExecContext(ctx,
		`INSERT INTO sys_migration (module, name) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		moduleName, name)
	return err
}

// RunModuleMigrations applies pending SQL files from addons/MODULE/migrations/*.sql.
func RunModuleMigrations(ctx context.Context, moduleName string) error {
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return fmt.Errorf("unknown module %q", moduleName)
	}
	dir := filepath.Join(addon.Path, "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)
	if len(files) == 0 {
		return nil
	}
	if err := EnsureMigrationTable(ctx); err != nil {
		return err
	}
	for _, name := range files {
		applied, err := migrationApplied(ctx, moduleName, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		sqlText := strings.TrimSpace(string(body))
		if sqlText == "" {
			continue
		}
		if _, err := orm.DB.ExecContext(ctx, sqlText); err != nil {
			return fmt.Errorf("migration %s/%s: %w", moduleName, name, err)
		}
		if err := markMigrationApplied(ctx, moduleName, name); err != nil {
			return err
		}
	}
	return nil
}

// RunAllMigrations applies migrations for every discovered addon in dependency order.
func RunAllMigrations(ctx context.Context) error {
	topo, err := sortAddonsTopo(DiscoveredAddons)
	if err != nil {
		return err
	}
	for _, a := range topo {
		if err := RunModuleMigrations(ctx, a.Manifest.Name); err != nil {
			return err
		}
	}
	return nil
}
