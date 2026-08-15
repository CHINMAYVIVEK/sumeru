package web_test

import (
	"testing"

	"sumeru/core/server/web"
)

func TestFlashFromQueryImportMessage(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("imported_5_updated_2_skipped_1")
	if !ok {
		t.Fatal("expected flash")
	}
	if flash.Kind != "success" || flash.Title != "Import complete" {
		t.Fatalf("flash = %+v", flash)
	}
}

func TestFlashFromQueryLegacyImport(t *testing.T) {
	flash, ok := web.FlashFromQueryMessage("imported_3")
	if !ok || flash.Body != "Imported 3 row(s)." {
		t.Fatalf("flash = %+v ok=%v", flash, ok)
	}
}
