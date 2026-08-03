// Package server loads configuration, runs persistence and module lifecycle, registers HTTP routes, and serves the app.
//
// Process entry should call Run. LoadConfig and AbsPaths use sumeru/core/server/config; InitDB,
// SyncModels, and module wiring delegate to sumeru/core/sdk where applicable.
//
// Addon models should import only sumeru/core/sdk (not sumeru/core/orm directly) so internal renames stay behind the sdk facade.
//
// Layout: sumeru/core/sdk (stable addon API), sumeru/core/orm (data layer), sumeru/core/engine (XML → UI),
// sumeru/core/server/web (HTTP), sumeru/core/server/router (route registry).
package server
