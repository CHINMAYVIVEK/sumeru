package module

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"sumeru/core/orm"
)

func sortAddonsTopo(discovered map[string]*Addon) ([]*Addon, error) {
	remainingAddons := make(map[string]*Addon)
	for addonName, addon := range discovered {
		remainingAddons[addonName] = addon
	}
	var sortedAddons []*Addon
	for len(remainingAddons) > 0 {
		var addonCandidates []string
		for name := range remainingAddons {
			addon := remainingAddons[name]
			isSatisfied := true
			for _, dependencyName := range addon.Manifest.Depends {
				dependencyName = strings.TrimSpace(dependencyName)
				if dependencyName == "" || dependencyName == name {
					continue
				}
				if _, has := discovered[dependencyName]; !has {
					continue
				}
				if !containsAddonName(sortedAddons, dependencyName) {
					isSatisfied = false
					break
				}
			}
			if isSatisfied {
				addonCandidates = append(addonCandidates, name)
			}
		}
		if len(addonCandidates) == 0 {
			keys := make([]string, 0, len(remainingAddons))
			for k := range remainingAddons {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			addonCandidates = keys[:1]
		}
		sort.Strings(addonCandidates)
		selectedAddonName := addonCandidates[0]
		sortedAddons = append(sortedAddons, remainingAddons[selectedAddonName])
		delete(remainingAddons, selectedAddonName)
	}
	return sortedAddons, nil
}

func containsAddonName(addonList []*Addon, addonName string) bool {
	for _, addon := range addonList {
		if addon.Manifest.Name == addonName {
			return true
		}
	}
	return false
}

// missingInstalledDependencies lists manifest depends that are not installed in sys.module
// (or not registered). On-disk deps missing from DiscoveredAddons are ignored.
func missingInstalledDependencies(context context.Context, moduleName string) ([]string, error) {
	addon, ok := DiscoveredAddons[moduleName]
	if !ok {
		return nil, nil
	}
	var missingDependencies []string
	for _, dependencyName := range addon.Manifest.Depends {
		dependencyName = strings.TrimSpace(dependencyName)
		if dependencyName == "" || dependencyName == moduleName {
			continue
		}
		if _, has := DiscoveredAddons[dependencyName]; !has {
			continue
		}
		moduleRow, err := moduleRow(context, dependencyName)
		if err != nil {
			if err == sql.ErrNoRows {
				missingDependencies = append(missingDependencies, dependencyName)
				continue
			}
			return nil, err
		}
		if moduleStateString(moduleRow) != "installed" {
			missingDependencies = append(missingDependencies, dependencyName)
		}
	}
	return missingDependencies, nil
}

func installedModuleDependingOn(context context.Context, targetModuleName string) (string, error) {
	moduleRows, err := orm.DB.QueryContext(context,
		`SELECT name, state FROM `+orm.GetTableName("sys.module")+` WHERE state = 'installed' AND name <> $1`,
		targetModuleName,
	)
	if err != nil {
		return "", err
	}
	defer moduleRows.Close()

	for moduleRows.Next() {
		var moduleName, state string
		if err := moduleRows.Scan(&moduleName, &state); err != nil {
			return "", err
		}
		addon, ok := DiscoveredAddons[moduleName]
		if !ok {
			continue
		}
		for _, dependencyName := range addon.Manifest.Depends {
			if strings.TrimSpace(dependencyName) == targetModuleName {
				return moduleName, nil
			}
		}
	}
	return "", moduleRows.Err()
}
