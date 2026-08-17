package module

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"sumeru/core/orm"
)

// ModuleInfo describes one addon on disk and optional DB state.
type ModuleInfo struct {
	Name        string
	Version     string
	State       string
	Depends     []string
	Application bool
}

// ListModuleInfo returns discovered addons sorted by name with installed state when DB is available.
func ListModuleInfo(ctx context.Context) ([]ModuleInfo, error) {
	names := make([]string, 0, len(DiscoveredAddons))
	for n := range DiscoveredAddons {
		names = append(names, n)
	}
	sort.Strings(names)
	states := map[string]string{}
	if orm.DB != nil && orm.IsInitialized() {
		rows, err := orm.DB.QueryContext(ctx,
			`SELECT name, state FROM `+orm.MustQuotedTableName("sys.module"))
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var name, state string
				if err := rows.Scan(&name, &state); err == nil {
					states[name] = state
				}
			}
		}
	}
	out := make([]ModuleInfo, 0, len(names))
	for _, n := range names {
		a := DiscoveredAddons[n]
		st := states[n]
		if st == "" {
			st = "uninstalled"
		}
		app := false
		if a.Manifest.Application != nil {
			app = *a.Manifest.Application
		}
		out = append(out, ModuleInfo{
			Name:        a.Manifest.Name,
			Version:     a.Manifest.Version,
			State:       st,
			Depends:     append([]string(nil), a.Manifest.Depends...),
			Application: app,
		})
	}
	return out, nil
}

// DependsTree prints dependency tree for moduleName.
func DependsTree(moduleName string) (string, error) {
	moduleName = strings.TrimSpace(moduleName)
	if moduleName == "" {
		return "", fmt.Errorf("module name required")
	}
	var b strings.Builder
	var walk func(name string, depth int, seen map[string]struct{}) error
	walk = func(name string, depth int, seen map[string]struct{}) error {
		if _, ok := seen[name]; ok {
			b.WriteString(strings.Repeat("  ", depth) + name + " (cycle)\n")
			return nil
		}
		seen[name] = struct{}{}
		a, ok := DiscoveredAddons[name]
		if !ok {
			b.WriteString(strings.Repeat("  ", depth) + name + " (missing)\n")
			return nil
		}
		b.WriteString(strings.Repeat("  ", depth) + name + "\n")
		for _, dep := range a.Manifest.Depends {
			dep = strings.TrimSpace(dep)
			if dep == "" {
				continue
			}
			childSeen := map[string]struct{}{}
			for k, v := range seen {
				childSeen[k] = v
			}
			if err := walk(dep, depth+1, childSeen); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(moduleName, 0, map[string]struct{}{}); err != nil {
		return "", err
	}
	return b.String(), nil
}
