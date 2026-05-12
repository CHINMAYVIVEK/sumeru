// Command sumeru-bp scaffolds a new addon under addons/ following strict Sumeru conventions.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sumeru/core/module"
)

func main() {
	name := flag.String("name", "", "technical module name (folder + manifest name), e.g. my_vendor_app")
	withModels := flag.Bool("with-models", false, "generate models/ package and blank-import it from init.go")
	outDir := flag.String("out", "addons", "parent directory for the new module (relative to repo root)")
	flag.Parse()

	if strings.TrimSpace(*name) == "" {
		fmt.Fprintln(os.Stderr, "usage: sumeru-bp -name my_module [-with-models] [-out addons]")
		os.Exit(2)
	}
	if !module.ModuleNamePattern.MatchString(*name) {
		fmt.Fprintf(os.Stderr, "invalid -name %q (must match %s)\n", *name, module.ModuleNamePattern.String())
		os.Exit(2)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	repoRoot, err := module.FindRepoRoot(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}
	modPath, err := module.ReadModulePath(repoRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %v\n", err)
		os.Exit(1)
	}

	target := filepath.Join(repoRoot, filepath.FromSlash(*outDir), *name)
	if fi, err := os.Stat(target); err == nil && fi.IsDir() {
		fmt.Fprintf(os.Stderr, "sumeru-bp: %q already exists\n", target)
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Join(target, "views"), 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	manifest := fmt.Sprintf(`{
  "name": %q,
  "version": "1.0.0",
  "depends": ["base"],
  "author": "Your Company",
  "description": "TODO: short summary for Apps.",
  "application": true,
  "data": [
    "views/menus.xml"
  ]
}
`, *name)
	if err := os.WriteFile(filepath.Join(target, "manifest.json"), []byte(manifest), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	menus := `<sumeru>
  <data>
    <menuitem id="menu_` + *name + `_root" name="` + humanTitle(*name) + `" sequence="10"/>
  </data>
</sumeru>
`
	if err := os.WriteFile(filepath.Join(target, "views", "menus.xml"), []byte(menus), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var initBody strings.Builder
	fmt.Fprintf(&initBody, "package %s\n\n", *name)
	if *withModels {
		modelImport := modPath + "/" + filepath.ToSlash(filepath.Join(*outDir, *name, "models"))
		fmt.Fprintf(&initBody, "import _ %q\n", modelImport)
		if err := os.MkdirAll(filepath.Join(target, "models"), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		tn := typeName(*name)
		modelDotted := strings.ReplaceAll(*name, "_", ".") + ".record"
		modelGo := fmt.Sprintf(`package models

import "sumeru/core/base"

type %s struct {
	base.BaseModel
	Name string `+"`db:\"name\"`"+`
}

func (m *%s) ModelName() string {
	return %q
}

func (m *%s) Fields() []base.FieldDefinition {
	return []base.FieldDefinition{
		{Name: "name", Type: base.Char, String: "Name", Required: true},
	}
}

func init() {
	base.RegisterModel(base.RegisterModelInput{Model: &%s{}})
}
`, tn, tn, modelDotted, tn, tn)
		if err := os.WriteFile(filepath.Join(target, "models", "models.go"), []byte(modelGo), 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := os.WriteFile(filepath.Join(target, "init.go"), []byte(initBody.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Created addon at %s\nNext: go generate ./cmd/sumeru && go run ./cmd/sumeru -- -c sumeru.conf\n", target)
}

func typeName(technical string) string {
	parts := strings.Split(technical, "_")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func humanTitle(technical string) string {
	parts := strings.Split(technical, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		low := strings.ToLower(p)
		parts[i] = strings.ToUpper(low[:1]) + low[1:]
	}
	return strings.Join(parts, " ")
}
