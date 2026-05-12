// Package server loads configuration, runs persistence and module lifecycle, registers HTTP routes, and serves the app.
//
// Process entry should call Run. LoadConfig and AbsPaths use sumeru/core/server/config; InitDB,
// SyncModels, and module wiring delegate to sumeru/core/base where applicable.
//
// Addon models should import only sumeru/core/base (not sumeru/core/orm directly) so internal renames stay behind the base facade.
//
// Layout: sumeru/core/base (stable addon API), sumeru/core/orm (data layer), sumeru/core/engine (XML → UI), sumeru/core/server/web (HTTP).
package server
