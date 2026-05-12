// Package platformmsg holds user-facing strings and log line formats shared by
// core/module (loader) and core/server/web. It lives under core/base/ to sit next
// to the filesystem "base" addon at core/base/base/ without importing sumeru/core/base
// (avoids import cycles: core/base imports core/module).
package platformmsg

// HTTP / HTML responses (web layer).
const (
	MsgHTTPTemplateError = "Template error"
)

// Log line formats (fmt.Printf / log.Printf style; end with \n where used as full lines).
const (
	FmtErrorSyncingAddon   = "Error syncing addon %s: %v\n"
	FmtLoadedAddonData     = "Loaded addon data: %s (v%s)\n"
	FmtAddonOverrideNotice = "Addon %q: %q overrides %q\n"
	FmtDataFileMissing     = "Warning: Data file %s not found in addon %s\n"
	FmtViewInheritWarning  = "Warning: view inherit in %s (record %s): %v\n"
	FmtGenericUpsertWarn   = "Warning: upsert %s record %s: %v\n"
)
