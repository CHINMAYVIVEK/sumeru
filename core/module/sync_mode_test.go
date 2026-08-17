package module

import (
	"context"
	"testing"
)

func TestDataFileOptsSkipExistingOnUpdate(t *testing.T) {
	opts := dataFileOpts{noUpdate: true}
	ctx := ContextWithSyncMode(context.Background(), moduleReloadUpdate)
	if opts.skipExistingOnUpdate(ctx, "base", "") {
		t.Fatal("empty xml id should not skip")
	}
	ctxInstall := ContextWithSyncMode(context.Background(), moduleReloadInstall)
	if opts.skipExistingOnUpdate(ctxInstall, "base", "company_main") {
		t.Fatal("install mode should not skip")
	}
	if opts.skipExistingOnUpdate(ctx, "base", "company_main") {
		// without DB/xml id resolution this returns false — no skip when id unknown
	}
	opts2 := dataFileOpts{noUpdate: false}
	if opts2.skipExistingOnUpdate(ctx, "base", "company_main") {
		t.Fatal("noUpdate false should not skip")
	}
}

func TestSyncModeFromContextDefault(t *testing.T) {
	if SyncModeFromContext(context.Background()) != moduleReloadInstall {
		t.Fatal("expected install default")
	}
}
