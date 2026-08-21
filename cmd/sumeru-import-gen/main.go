// Command sumeru-import-gen writes a Go file of blank imports for all strict addons
// discovered via INI addons_path (default: cmd/sumeru/zimports.go, package main for in-repo builds).
// For an external workspace (e.g. sumeru_custom_addons), use -workspace, -package and -out to write
// only under the workspace tree (init.go, zmodels.go, zrefs.go, addonimports).
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/importgen"
	"sumeru/core/module"
)

func main() {
	outPath := flag.String("out", "cmd/sumeru/zimports.go", "output .go path: relative to -root or -workspace, or absolute")
	configPath := flag.String("config", "sumeru.conf.example", "INI path: relative to repo/workspace root, or absolute")
	repoRootFlag := flag.String("root", "", "sumeru repo root (directory with go.mod); default: search upward from cwd")
	workspaceFlag := flag.String("workspace", "", "external workspace root (e.g. sumeru_custom_addons); limits writes to this tree")
	addonsRootFlag := flag.String("addons", "", "standard addons root for read-only zrefs scan (e.g. ../sumeru_addons)")
	packageName := flag.String("package", "main", "generated file package name (e.g. main or addonimports)")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	workspace := strings.TrimSpace(*workspaceFlag)
	if workspace != "" {
		workspace, err = filepath.Abs(workspace)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		sumeruRoot := strings.TrimSpace(*repoRootFlag)
		if sumeruRoot == "" {
			fmt.Fprintln(os.Stderr, "sumeru-import-gen: -workspace requires -root (sumeru engine path)")
			os.Exit(1)
		}
		sumeruRoot, err = filepath.Abs(sumeruRoot)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		addonsRoot := strings.TrimSpace(*addonsRootFlag)
		dest, err := importgen.RunWorkspaceGen(workspace, sumeruRoot, addonsRoot, *configPath, *outPath, *packageName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sumeru-import-gen: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(dest)
		return
	}

	repoRoot := strings.TrimSpace(*repoRootFlag)
	if repoRoot == "" {
		repoRoot, err = module.FindRepoRoot(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sumeru-import-gen: %v (use -root)\n", err)
			os.Exit(1)
		}
	} else {
		repoRoot, err = filepath.Abs(repoRoot)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	dest, err := importgen.RunGen(repoRoot, *configPath, *outPath, *packageName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-import-gen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(dest)
}
