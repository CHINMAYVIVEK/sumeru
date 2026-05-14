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

// ReadModulePath returns the Go module path from go.mod in repoRoot (e.g. "sumeru").
func ReadModulePath(repoRoot string) (string, error) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module directive in %s", filepath.Join(repoRoot, "go.mod"))
}

// FindRepoRoot walks parents from any path under the repo until go.mod is found.
func FindRepoRoot(fromAbs string) (string, error) {
	dir := fromAbs
	if fi, err := os.Stat(dir); err == nil && !fi.IsDir() {
		dir = filepath.Dir(dir)
	}
	for i := 0; i < 20; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found above %q", fromAbs)
}

func isBuiltinAddonPath(addonPath string) bool {
	s := filepath.Clean(addonPath)
	if strings.Contains(s, "module"+string(filepath.Separator)+"builtin") {
		return true
	}
	// .../core/base/base — minimal "base" sys.module (manifest only; no init.go requirement)
	return strings.Contains(filepath.ToSlash(s), "/core/base/base") &&
		filepath.Base(s) == "base" && filepath.Base(filepath.Dir(s)) == "base"
}

// addonGoModuleContext resolves the Go module root, module import path, and repo-relative
// directory path for an addon filesystem path.
func addonGoModuleContext(addonPath string) (repoRoot, modPath, rel string, err error) {
	addonPath = filepath.Clean(addonPath)
	repoRoot, err = FindRepoRoot(addonPath)
	if err != nil {
		return "", "", "", err
	}
	modPath, err = ReadModulePath(repoRoot)
	if err != nil {
		return "", "", "", err
	}
	rel, err = filepath.Rel(repoRoot, addonPath)
	if err != nil {
		return "", "", "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", "", "", fmt.Errorf("addon path %q escapes module root %q", addonPath, repoRoot)
	}
	return repoRoot, modPath, rel, nil
}

// ValidateDiscoveredAddons checks strict layout for every discovered addon.
// Filesystem-only modules under core/module/builtin or core/base/base skip the Go init.go / models rules.
// Each addon is validated against the Go module root that contains it (supports multiple
// addon roots, e.g. standard sumeru plus a sibling workspace module).
func ValidateDiscoveredAddons(discovered map[string]*Addon) error {
	if len(discovered) == 0 {
		return nil
	}

	names := make([]string, 0, len(discovered))
	for n := range discovered {
		names = append(names, n)
	}
	sort.Strings(names)

	var errs []string
	for _, name := range names {
		a := discovered[name]
		if err := validateOneAddon(a); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", a.Manifest.Name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("addon convention violations:\n- %s", strings.Join(errs, "\n- "))
	}
	return nil
}

func validateOneAddon(a *Addon) error {
	m := &a.Manifest
	dirName := filepath.Base(a.Path)

	if !ModuleNamePattern.MatchString(m.Name) {
		return fmt.Errorf("manifest name %q must match %s", m.Name, ModuleNamePattern.String())
	}
	if m.Name != dirName {
		return fmt.Errorf("folder name %q must equal manifest name %q", dirName, m.Name)
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("manifest version is required")
	}
	if strings.TrimSpace(m.Author) == "" {
		return fmt.Errorf("manifest author is required")
	}
	if strings.TrimSpace(m.Description) == "" {
		return fmt.Errorf("manifest description is required")
	}
	dataFiles := m.Data
	if dataFiles == nil {
		dataFiles = []string{}
	}

	for _, rel := range dataFiles {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			return fmt.Errorf("manifest data contains empty entry")
		}
		fp := filepath.Join(a.Path, rel)
		if _, err := os.Stat(fp); err != nil {
			return fmt.Errorf("data file %q: %w", rel, err)
		}
		if strings.HasSuffix(strings.ToLower(rel), ".xml") {
			if err := validateModuleXMLRoot(fp); err != nil {
				return fmt.Errorf("data %q: %w", rel, err)
			}
		}
	}

	if isBuiltinAddonPath(a.Path) {
		return nil
	}

	initPath := filepath.Join(a.Path, "init.go")
	initBytes, err := os.ReadFile(initPath)
	if err != nil {
		return fmt.Errorf("strict addon requires init.go: %w", err)
	}
	if err := validateRootInitGo(string(initBytes), m.Name); err != nil {
		return err
	}

	modelsDir := filepath.Join(a.Path, "models")
	hasModels, err := dirHasGoFiles(modelsDir)
	if err != nil {
		return err
	}
	if hasModels {
		want, err := expectedModelsImport(a.Path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(initBytes), `"`+want+`"`) {
			return fmt.Errorf("init.go must blank-import models package %q (found models/*.go)", want)
		}
	}

	return nil
}

func validateRootInitGo(src, wantPkg string) error {
	lines := strings.Split(src, "\n")
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
			if fields[1] != wantPkg {
				return fmt.Errorf("init.go package must be %q, got %q", wantPkg, fields[1])
			}
			return nil
		}
	}
	return fmt.Errorf("init.go: missing package declaration")
}

func expectedModelsImport(addonPath string) (string, error) {
	_, modPath, rel, err := addonGoModuleContext(addonPath)
	if err != nil {
		return "", err
	}
	return modPath + "/" + filepath.ToSlash(rel) + "/models", nil
}

func dirHasGoFiles(dir string) (bool, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if !fi.IsDir() {
		return false, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".go") && !strings.HasSuffix(e.Name(), "_test.go") {
			return true, nil
		}
	}
	return false, nil
}

func validateModuleXMLRoot(xmlPath string) error {
	b, err := os.ReadFile(xmlPath)
	if err != nil {
		return err
	}
	s := string(bytes.TrimPrefix(b, []byte{0xEF, 0xBB, 0xBF}))
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "<?xml") {
		if i := strings.Index(s, ">"); i >= 0 {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	if !strings.HasPrefix(s, "<sumeru") {
		return fmt.Errorf("module XML root must be <sumeru>, file does not start with <sumeru")
	}
	return nil
}

// AddonImportPaths returns blank-import paths for every strict addon that ships Go (init.go).
// Each path is under its own Go module root (e.g. sumeru/addons/sales and
// sumeru_custom_addons/addons/acme_demo for a workspace sibling).
func AddonImportPaths(discovered map[string]*Addon) ([]string, error) {
	var out []string
	names := make([]string, 0, len(discovered))
	for n := range discovered {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		a := discovered[name]
		if isBuiltinAddonPath(a.Path) {
			continue
		}
		initPath := filepath.Join(a.Path, "init.go")
		if _, err := os.Stat(initPath); err != nil {
			continue
		}
		_, modPath, rel, err := addonGoModuleContext(a.Path)
		if err != nil {
			return nil, fmt.Errorf("addon %s: %w", a.Manifest.Name, err)
		}
		out = append(out, modPath+"/"+filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out, nil
}
