// Command sumeru-import-gen writes a Go file of blank imports for all strict addons
// discovered via INI addons_path (default: cmd/sumeru/zimports.go, package main for in-repo builds).
// For an external workspace (e.g. sumeru_custom_addons), use -package and -out to write
// only side-effect imports outside the standard sumeru tree.
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
	outPath := flag.String("out", "cmd/sumeru/zimports.go", "output .go path: relative to -root/repo, or absolute")
	configPath := flag.String("config", "sumeru.conf.example", "INI path: relative to repo root, or absolute")
	repoRootFlag := flag.String("root", "", "sumeru repo root (directory with go.mod); default: search upward from cwd")
	packageName := flag.String("package", "main", "generated file package name (e.g. main or addonimports)")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
