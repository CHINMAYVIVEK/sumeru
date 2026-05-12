// Package base is the stable surface for addon and integration code.
//
// Import sumeru/core/base (not sumeru/core/orm or sumeru/core/server directly) so renames
// in engine, orm, or server can be absorbed here without breaking downstream modules.
//
// Default filesystem modules shipped with the repo live under core/base/user,
// core/base/company, etc. (see core/base/base for the technical "base" app).
//
// Public API (prefer struct inputs):
//   - Types: Model, FieldDefinition, FieldType, BaseModel; field-type constants (Char, Text, …).
//   - RegisterModel(RegisterModelInput)
//   - Cfg() (INI is loaded by the process entrypoint via sumeru/core/server/config)
//   - InitDB(InitDBInput), SyncModels(SyncModelsInput)
//   - LoadAddonPaths(LoadAddonPathsInput), RunModuleCLI(RunModuleCLIInput)
//   - SetExtraStylesheetURLs(SetExtraStylesheetURLsInput)
//   - SearchOne, Search, Create, Upsert, ResolveXmlId (each with corresponding *Input struct)
//   - AsString, CoerceInt64 (*Input structs)
package base
