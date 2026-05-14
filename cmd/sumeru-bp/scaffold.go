package main

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

type ScaffoldConfig struct {
	Name         string
	HumanTitle   string
	TypeName     string
	ModelDotted  string
	ModelImport  string
	OutDir       string
	RepoRoot     string
	ModPath      string
}

func RunScaffold(cfg *ScaffoldConfig) {
	target := filepath.Join(cfg.RepoRoot, filepath.FromSlash(cfg.OutDir), cfg.Name)
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %q already exists\n", target)
		os.Exit(1)
	}

	// Create directory structure
	dirs := []string{
		"models",
		"views",
		"security",
		"static/src/img",
		"static/src/css",
		"static/src/js",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(target, filepath.FromSlash(d)), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Render templates
	renderTemplate(cfg, "manifest.json.tmpl", filepath.Join(target, "manifest.json"))
	renderTemplate(cfg, "init.go.tmpl", filepath.Join(target, "init.go"))
	renderTemplate(cfg, "models.go.tmpl", filepath.Join(target, "models", "models.go"))
	
	// Security
	renderTemplate(cfg, "security.xml.tmpl", filepath.Join(target, "security", "security.xml"))
	renderTemplate(cfg, "sys.access.csv.tmpl", filepath.Join(target, "security", "sys.access.csv"))

	// Views & Actions
	renderTemplate(cfg, "actions.xml.tmpl", filepath.Join(target, "views", "actions.xml"))
	renderTemplate(cfg, "form_view.xml.tmpl", filepath.Join(target, "views", "form_view.xml"))
	renderTemplate(cfg, "tree_view.xml.tmpl", filepath.Join(target, "views", "tree_view.xml"))
	renderTemplate(cfg, "kanban_view.xml.tmpl", filepath.Join(target, "views", "kanban_view.xml"))
	renderTemplate(cfg, "menus.xml.tmpl", filepath.Join(target, "views", "menus.xml"))

	fmt.Printf("Successfully created premium modular addon at: %s\n", target)
	fmt.Printf("Next steps:\n")
	if cfg.ModPath == "sumeru_custom_addons" {
		fmt.Printf("  1. Run 'make generate' in the workspace\n")
		fmt.Printf("  2. Run './sumeru-workspace.sh -i %s'\n", cfg.Name)
	} else {
		fmt.Printf("  1. Run 'make generate'\n")
		fmt.Printf("  2. Run 'go run ./cmd/sumeru -- -i %s'\n", cfg.Name)
	}
}

func renderTemplate(cfg *ScaffoldConfig, tmplName, outPath string) {
	t, err := template.ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing template %s: %v\n", tmplName, err)
		os.Exit(1)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error executing template %s: %v\n", tmplName, err)
		os.Exit(1)
	}

	writeOrDie(outPath, buf.String())
}
