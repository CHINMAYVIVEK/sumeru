// Package sdk is the stable surface for addon and integration code.
//
// Import sumeru/core/sdk (not sumeru/core/orm or sumeru/core/server directly) so renames
// in engine, orm, or server can be absorbed here without breaking downstream modules.
//
// Kernel models (core.user, core.company, …) register from addons/base/models (manifest "base").
// The Go package name is sdk; the installable addon technical name remains "base"
// (package base under addons/base — do not confuse the two).
//
// Public API (prefer struct inputs):
//   - Types: Model, FieldDefinition, FieldType, BaseModel; field-type constants (Char, Text, …).
//   - RegisterModel(RegisterModelInput)
//   - Cfg() (INI is loaded by the process entrypoint via sumeru/core/server/config)
//   - InitDB(InitDBInput), SyncModels(SyncModelsInput)
//   - SearchOne, Search, Create, Upsert, ResolveXmlId (each with corresponding *Input struct)
//   - ExportCSV, ExportPDF, BulkTemplateCSV, PreviewBulkImport, ExecuteBulkImport (see report.go)
//   - RegisterReportCellFormatter for custom export cell text
//
// Module load / CLI and stylesheet registration live on sumeru/core/server (and module/render)
// so this package stays importable from addon models without cycles into engine or addons.
package sdk
