package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ScssBundle handles the discovery and ordering of SCSS files from core and addons.
type ScssBundle struct {
	Overrides []string // Addon variable overrides
	CoreBase  []string // Core variables and layout
	AddonStyles []string // Addon specific styles
}

// DiscoverSCSS scans all active addons for SCSS files and prepares the bundle order.
func (b *ScssBundle) DiscoverSCSS(corePath string, addonPaths map[string]string) {
	// 1. Add Core Variables first in CoreBase (but they will be imported after overrides)
	b.CoreBase = append(b.CoreBase, filepath.Join(corePath, "engine/assets/scss/variables"))
	
	// 2. Discover Addon Overrides
	for _, path := range addonPaths {
		overridePath := filepath.Join(path, "static/scss/overrides")
		if _, err := os.Stat(overridePath + ".scss"); err == nil {
			b.Overrides = append(b.Overrides, overridePath)
		}
	}

	// 3. Add Core Layouts/Components
	b.CoreBase = append(b.CoreBase, filepath.Join(corePath, "engine/assets/scss/base"))
	b.CoreBase = append(b.CoreBase, filepath.Join(corePath, "engine/assets/scss/shell"))
	b.CoreBase = append(b.CoreBase, filepath.Join(corePath, "engine/assets/scss/views"))
	b.CoreBase = append(b.CoreBase, filepath.Join(corePath, "engine/assets/scss/compat"))

	// 4. Discover Addon Specific Styles
	for _, path := range addonPaths {
		stylePath := filepath.Join(path, "static/scss/style")
		if _, err := os.Stat(stylePath + ".scss"); err == nil {
			b.AddonStyles = append(b.AddonStyles, stylePath)
		}
	}
}

// GenerateManifest creates a single entry point .scss file that imports everything in order.
func (b *ScssBundle) GenerateManifest() string {
	var sb strings.Builder
	sb.WriteString("// Sumeru Generated SCSS Manifest\n")
	sb.WriteString("// DO NOT EDIT MANUALLY - Managed by Sumeru Engine\n\n")

	sb.WriteString("// 1. Addon Overrides\n")
	for _, p := range b.Overrides {
		sb.WriteString(fmt.Sprintf("@import \"%s\";\n", b.toSassPath(p)))
	}

	sb.WriteString("\n// 2. Core Base (Variables & Layout)\n")
	for _, p := range b.CoreBase {
		sb.WriteString(fmt.Sprintf("@import \"%s\";\n", b.toSassPath(p)))
	}

	sb.WriteString("\n// 3. Addon Styles\n")
	for _, p := range b.AddonStyles {
		sb.WriteString(fmt.Sprintf("@import \"%s\";\n", b.toSassPath(p)))
	}

	return sb.String()
}

func (b *ScssBundle) toSassPath(p string) string {
	// Sass imports should use forward slashes even on Windows
	return strings.ReplaceAll(p, "\\", "/")
}

// RebuildSCSSManifest coordinates the full discovery and generation process.
func RebuildSCSSManifest(corePath string, addonPaths map[string]string) error {
	bundle := ScssBundle{}
	bundle.DiscoverSCSS(corePath, addonPaths)
	manifest := bundle.GenerateManifest()

	outputPath := filepath.Join(corePath, "engine/assets/scss/main.scss")
	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(manifest), 0644)
}
