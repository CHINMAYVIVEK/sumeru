package importgen

import (
	"fmt"
	"os"
	"path/filepath"

	"sumeru/core/module"
	"sumeru/core/server/config"
)

type pipelineResult struct {
	dest       string
	discovered map[string]*module.Addon
}

func loadAndDiscover(root, configPath, outPath, packageName string) (*pipelineResult, error) {
	if err := ValidateOutputPackage(packageName); err != nil {
		return nil, err
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	dest, err := ResolveOutputPath(root, outPath)
	if err != nil {
		return nil, err
	}

	cfg := configPathForLoad(root, configPath)
	cfgAbs, err := filepath.Abs(cfg)
	if err != nil {
		return nil, fmt.Errorf("config path: %w", err)
	}
	if err := config.LoadConfig(cfgAbs); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err := config.AbsPaths(); err != nil {
		return nil, fmt.Errorf("abs paths: %w", err)
	}

	discovered, err := module.DiscoverAddonRoots(config.AppConfig.AddonPaths)
	if err != nil {
		return nil, fmt.Errorf("discover: %w", err)
	}
	return &pipelineResult{dest: dest, discovered: discovered}, nil
}

func writeImportFile(dest, packageName string, discovered map[string]*module.Addon) error {
	imports, err := module.AddonImportPaths(discovered)
	if err != nil {
		return err
	}
	body := BuildImportGoFile(packageName, imports)
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	if err := os.WriteFile(dest, []byte(body), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", dest, err)
	}
	return nil
}
