package module

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/applog"
	"sumeru/core/orm"
)

// LoadAddonPaths discovers addons from multiple roots (comma-separated in config),
// syncs sys.module rows (if DB initialized), then loads XML for installed & active modules.
func LoadAddonPaths(rootPaths []string) error {
	contextWithBypass := orm.ContextWithBypass(context.Background(), true)

	var sanitizedRoots []string
	for _, path := range rootPaths {
		if strings.TrimSpace(path) != "" {
			sanitizedRoots = append(sanitizedRoots, strings.TrimSpace(path))
		}
	}
	if len(sanitizedRoots) == 0 {
		return fmt.Errorf("no addon roots configured")
	}

	discoveredAddons, err := DiscoverAddonRoots(sanitizedRoots)
	if err != nil {
		return err
	}
	if err := ValidateDiscoveredAddons(discoveredAddons); err != nil {
		return err
	}
	DiscoveredAddons = discoveredAddons
	for addonName, addon := range discoveredAddons {
		LoadedAddons[addonName] = addon
	}

	if orm.IsInitialized() {
		if err := syncSysModuleRows(contextWithBypass, discoveredAddons); err != nil {
			return err
		}

		sortedAddonOrder, err := sortAddonsTopo(discoveredAddons)
		if err != nil {
			return err
		}

		var syncErrs []error
		for _, addon := range sortedAddonOrder {
			// Only load XML data for modules that are installed and active.
			// Discovered-but-uninstalled modules are visible in the Apps UI
			// but their menus, views, and records are not loaded until installed.
			if !shouldSyncData(contextWithBypass, addon.Manifest.Name) {
				continue
			}
			if err := addon.SyncToDB(contextWithBypass); err != nil {
				if IsFatalSync(err) {
					_ = setModuleLastError(contextWithBypass, addon.Manifest.Name, err.Error())
					syncErrs = append(syncErrs, err)
					applog.WarnMsg(contextWithBypass, "module", "sync",
						fmt.Sprintf("Error syncing addon %s", addon.Manifest.Name), err, nil)
					continue
				}
				applog.WarnMsg(contextWithBypass, "module", "sync",
					fmt.Sprintf("Error syncing addon %s", addon.Manifest.Name), err, nil)
			} else {
				_ = setModuleLastError(contextWithBypass, addon.Manifest.Name, "")
				applog.InfoMsg(contextWithBypass, "module", "sync",
					fmt.Sprintf("Loaded addon data: %s (v%s)", addon.Manifest.Name, addon.Manifest.Version), nil)
			}
		}
		if len(syncErrs) > 0 {
			return fmt.Errorf("fatal module sync failure(s): %w", syncErrs[0])
		}
	} else {
		applog.InfoMsg(contextWithBypass, "module", "discover",
			"Database not initialized; skipping DB sync for discovered addons",
			map[string]interface{}{"count": len(discoveredAddons)})
	}

	return nil
}

// DiscoverAddonRoots scans each root for subdirectories containing manifest.json
// and returns the merged map keyed by manifest name (later roots override earlier).
func DiscoverAddonRoots(rootPaths []string) (map[string]*Addon, error) {
	addonsMap := map[string]*Addon{}
	for _, root := range rootPaths {
		dirEntries, err := os.ReadDir(root)
		if err != nil {
			return nil, fmt.Errorf("addons root %q: %w", root, err)
		}
		for _, dirEntry := range dirEntries {
			if !dirEntry.IsDir() {
				continue
			}
			addonPath := filepath.Join(root, dirEntry.Name())
			manifestPath := filepath.Join(addonPath, "manifest.json")
			if _, err := os.Stat(manifestPath); err != nil {
				continue
			}
			parsedManifest, err := parseManifest(manifestPath)
			if err != nil {
				return nil, fmt.Errorf("%s: manifest.json: %w", addonPath, err)
			}
			if existingAddon, exists := addonsMap[parsedManifest.Name]; exists {
				applog.InfoMsg(context.Background(), "module", "discover",
					fmt.Sprintf("Addon %q path override", parsedManifest.Name),
					map[string]interface{}{"new_path": addonPath, "old_path": existingAddon.Path})
			}
			addonsMap[parsedManifest.Name] = &Addon{Manifest: *parsedManifest, Path: addonPath}
		}
	}
	return addonsMap, nil
}

func parseManifest(filePath string) (*Manifest, error) {
	manifestData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var parsedManifest Manifest
	if err := json.Unmarshal(manifestData, &parsedManifest); err != nil {
		return nil, err
	}

	return &parsedManifest, nil
}

func displayNameForAddon(addonName string) string {
	if addonName == "" {
		return addonName
	}
	switch addonName {
	case "company":
		return "Companies"
	case "user":
		return "Users"
	case "base":
		return "General Settings"
	default:
		return strings.ToUpper(addonName[:1]) + addonName[1:]
	}
}

func irModuleDisplayName(addon *Addon) string {
	if displayName := strings.TrimSpace(addon.Manifest.DisplayName); displayName != "" {
		return displayName
	}
	return displayNameForAddon(addon.Manifest.Name)
}
