package orm

import (
	"slices"
	"testing"
)

func TestModelsForModuleSchemaSyncBaseIncludesTranslation(t *testing.T) {
	Registry = map[string]Model{
		"sys.module":      stubModel{name: "sys.module", fields: []FieldDefinition{{Name: "name", Type: Char}}},
		"sys.translation": stubModel{name: "sys.translation", fields: []FieldDefinition{{Name: "lang", Type: Char}}},
	}
	modelDeclaringModule = map[string]string{
		"sys.module":      "base",
		"sys.translation": "base",
	}
	t.Cleanup(func() {
		Registry = map[string]Model{}
		modelDeclaringModule = map[string]string{}
	})

	names, scoped := ModelsForModuleSchemaSync("base")
	if !scoped {
		t.Fatal("expected scoped sync for base")
	}
	if !slices.Contains(names, "sys.translation") {
		t.Fatalf("expected sys.translation in base schema sync, got %v", names)
	}
}
