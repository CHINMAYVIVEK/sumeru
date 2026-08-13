// Package platformmsg holds user-facing strings and log line formats shared by
// core/module (loader) and core/server/web. It lives under core/sdk/platformmsg
// without importing the sdk facade (avoids import cycles: sdk may reach module).
package platformmsg

// HTTP / HTML responses (web layer).
const (
	MsgHTTPTemplateError = "Template error"
)

// Message format strings for sync warnings (used with applog via syncWarn).
const (
	FmtGenericUpsertWarn = "Warning: upsert %s record %s: %v\n"
)
