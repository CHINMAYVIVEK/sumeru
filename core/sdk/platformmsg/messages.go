// Package platformmsg holds user-facing strings and log line formats shared by
// core/module (loader) and core/server/web. It lives under core/sdk/platformmsg
// without importing the sdk facade (avoids import cycles: sdk may reach module).
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
