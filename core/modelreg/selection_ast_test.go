package modelreg

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSelectionConstsFromSourcePackage(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller")
	}
	pkgDir := filepath.Join(filepath.Dir(file), "testdata", "selection")
	opts := selectionOptionsFromPackage(pkgDir, "Priority")
	if len(opts) != 3 {
		t.Fatalf("expected 3 priority options, got %v", opts)
	}
}
