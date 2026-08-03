package module

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ModuleNamePattern is the strict technical name for addon directories and manifest "name".
var ModuleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// ReadModulePath returns the Go module path from go.mod in repositoryRoot (e.g. "sumeru").
func ReadModulePath(repositoryRoot string) (string, error) {
	moduleData, err := os.ReadFile(filepath.Join(repositoryRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(moduleData), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module directive in %s", filepath.Join(repositoryRoot, "go.mod"))
}

// FindRepoRoot walks parents from any path under the repo until go.mod is found.
func FindRepoRoot(fromAbsolutePath string) (string, error) {
	currentDirectory := fromAbsolutePath
	if fileInfo, err := os.Stat(currentDirectory); err == nil && !fileInfo.IsDir() {
		currentDirectory = filepath.Dir(currentDirectory)
	}
	for i := 0; i < 20; i++ {
		if _, err := os.Stat(filepath.Join(currentDirectory, "go.mod")); err == nil {
			return currentDirectory, nil
		}
		parentDirectory := filepath.Dir(currentDirectory)
		if parentDirectory == currentDirectory {
			break
		}
		currentDirectory = parentDirectory
	}
	return "", fmt.Errorf("go.mod not found above %q", fromAbsolutePath)
}

func isBuiltinAddonPath(addonPath string) bool {
	sanitizedPath := filepath.Clean(addonPath)
	return strings.Contains(sanitizedPath, "module"+string(filepath.Separator)+"builtin")
}

// addonGoModuleContext resolves the Go module root, module import path, and repo-relative
// directory path for an addon filesystem path.
func addonGoModuleContext(addonPath string) (repositoryRoot, moduleImportPath, relativePath string, err error) {
	addonPath = filepath.Clean(addonPath)
	repositoryRoot, err = FindRepoRoot(addonPath)
	if err != nil {
		return "", "", "", err
	}
	moduleImportPath, err = ReadModulePath(repositoryRoot)
	if err != nil {
		return "", "", "", err
	}
	relativePath, err = filepath.Rel(repositoryRoot, addonPath)
	if err != nil {
		return "", "", "", err
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		return "", "", "", fmt.Errorf("addon path %q escapes module root %q", addonPath, repositoryRoot)
	}
	return repositoryRoot, moduleImportPath, relativePath, nil
}

// ValidateDiscoveredAddons checks strict layout for every discovered addon.
// Filesystem-only modules under core/module/builtin skip the Go init.go / models rules.
// Each addon is validated against the Go module root that contains it (supports multiple
// addon roots, e.g. standard sumeru plus a sibling workspace module).
func ValidateDiscoveredAddons(discoveredAddons map[string]*Addon) error {
	if len(discoveredAddons) == 0 {
		return nil
	}

	addonNames := make([]string, 0, len(discoveredAddons))
	for name := range discoveredAddons {
		addonNames = append(addonNames, name)
	}
	sort.Strings(addonNames)

	var validationErrors []string
	for _, name := range addonNames {
		addon := discoveredAddons[name]
		if err := validateOneAddon(addon); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("%s: %v", addon.Manifest.Name, err))
		}
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("addon convention violations:\n- %s", strings.Join(validationErrors, "\n- "))
	}
	return nil
}

func validateOneAddon(addon *Addon) error {
	manifest := &addon.Manifest
	directoryName := filepath.Base(addon.Path)

	if !ModuleNamePattern.MatchString(manifest.Name) {
		return fmt.Errorf("manifest name %q must match %s", manifest.Name, ModuleNamePattern.String())
	}
	if manifest.Name != directoryName {
		return fmt.Errorf("folder name %q must equal manifest name %q", directoryName, manifest.Name)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	if strings.TrimSpace(manifest.Author) == "" {
		return fmt.Errorf("manifest author is required")
	}
	if strings.TrimSpace(manifest.Description) == "" {
		return fmt.Errorf("manifest description is required")
	}
	dataFiles := manifest.Data
	if dataFiles == nil {
		dataFiles = []string{}
	}

	for _, relativeDataPath := range dataFiles {
		relativeDataPath = strings.TrimSpace(relativeDataPath)
		if relativeDataPath == "" {
			return fmt.Errorf("manifest data contains empty entry")
		}
		absoluteFilePath := filepath.Join(addon.Path, relativeDataPath)
		if _, err := os.Stat(absoluteFilePath); err != nil {
			return fmt.Errorf("data file %q: %w", relativeDataPath, err)
		}
		if strings.HasSuffix(strings.ToLower(relativeDataPath), ".xml") {
			if err := validateModuleXMLRoot(absoluteFilePath); err != nil {
				return fmt.Errorf("data %q: %w", relativeDataPath, err)
			}
		}
	}

	if isBuiltinAddonPath(addon.Path) {
		return nil
	}

	initPath := filepath.Join(addon.Path, "init.go")
	initFileBytes, err := os.ReadFile(initPath)
	if err != nil {
		return fmt.Errorf("strict addon requires init.go: %w", err)
	}
	if err := validateRootInitGo(string(initFileBytes), manifest.Name); err != nil {
		return err
	}

	modelsDirectory := filepath.Join(addon.Path, "models")
	hasModels, err := dirHasGoFiles(modelsDirectory)
	if err != nil {
		return err
	}
	if hasModels {
		expectedImport, err := expectedModelsImport(addon.Path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(initFileBytes), `"`+expectedImport+`"`) {
			return fmt.Errorf("init.go must blank-import models package %q (found models/*.go)", expectedImport)
		}
	}

	return nil
}

func validateRootInitGo(sourceCode, expectedPackage string) error {
	lines := strings.Split(sourceCode, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "package ") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				return fmt.Errorf("invalid package declaration in init.go")
			}
			if fields[1] != expectedPackage {
				return fmt.Errorf("init.go package must be %q, got %q", expectedPackage, fields[1])
			}
			return nil
		}
	}
	return fmt.Errorf("init.go: missing package declaration")
}

func expectedModelsImport(addonPath string) (string, error) {
	_, moduleImportPath, relativePath, err := addonGoModuleContext(addonPath)
	if err != nil {
		return "", err
	}
	return moduleImportPath + "/" + filepath.ToSlash(relativePath) + "/models", nil
}

func dirHasGoFiles(directoryPath string) (bool, error) {
	fileInfo, err := os.Stat(directoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !fileInfo.IsDir() {
		return false, nil
	}
	directoryEntries, err := os.ReadDir(directoryPath)
	if err != nil {
		return false, err
	}
	for _, entry := range directoryEntries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			return true, nil
		}
	}
	return false, nil
}

func validateModuleXMLRoot(xmlFilePath string) error {
	fileBytes, err := os.ReadFile(xmlFilePath)
	if err != nil {
		return err
	}
	contentString := string(bytes.TrimPrefix(fileBytes, []byte{0xEF, 0xBB, 0xBF}))
	contentString = strings.TrimSpace(contentString)
	if strings.HasPrefix(contentString, "<?xml") {
		if index := strings.Index(contentString, ">"); index >= 0 {
			contentString = strings.TrimSpace(contentString[index+1:])
		}
	}
	if !strings.HasPrefix(contentString, "<sumeru") {
		return fmt.Errorf("module XML root must be <sumeru>, file does not start with <sumeru")
	}
	return nil
}

// AddonImportPaths returns blank-import paths for every strict addon that ships Go (init.go).
// Each path is under its own Go module root (e.g. sumeru/addons/sales and
// sumeru_custom_addons/addons/acme_demo for a workspace sibling).
func AddonImportPaths(discoveredAddons map[string]*Addon) ([]string, error) {
	var importPaths []string
	addonNames := make([]string, 0, len(discoveredAddons))
	for name := range discoveredAddons {
		addonNames = append(addonNames, name)
	}
	sort.Strings(addonNames)
	for _, name := range addonNames {
		addon := discoveredAddons[name]
		if isBuiltinAddonPath(addon.Path) {
			continue
		}
		initPath := filepath.Join(addon.Path, "init.go")
		if _, err := os.Stat(initPath); err != nil {
			continue
		}
		_, moduleImportPath, relativePath, err := addonGoModuleContext(addon.Path)
		if err != nil {
			return nil, fmt.Errorf("addon %s: %w", addon.Manifest.Name, err)
		}
		importPaths = append(importPaths, moduleImportPath+"/"+filepath.ToSlash(relativePath))
	}
	sort.Strings(importPaths)
	return importPaths, nil
}
