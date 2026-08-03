package server

import (
	"sumeru/core/engine/render"
	"sumeru/core/module"
	"sumeru/core/sdk"
	"sumeru/core/server/config"
)

// Config is the on-disk configuration shape (stable alias via sdk).
type Config = sdk.Config

// Cfg returns a pointer to the live process configuration.
func Cfg() *config.Config {
	return sdk.Cfg()
}

// LoadConfig loads INI configuration from path.
func LoadConfig(path string) error {
	return config.LoadConfig(path)
}

// AbsPaths resolves relative paths in AppConfig.
func AbsPaths() error {
	return config.AbsPaths()
}

// InitDB opens the database pool.
func InitDB(databaseConnectionString string) {
	sdk.InitDB(sdk.InitDBInput{DSN: databaseConnectionString})
}

// SyncModels syncs registered models to the database.
func SyncModels() error {
	return sdk.SyncModels(sdk.SyncModelsInput{})
}

// LoadAddonPaths discovers and loads addons from the given roots.
func LoadAddonPaths(paths []string) error {
	return module.LoadAddonPaths(paths)
}

// RunModuleCLI runs -i / -u module lists.
func RunModuleCLI(installCSV, updateCSV string) error {
	return module.RunModuleCLI(installCSV, updateCSV)
}

// SetExtraStylesheetURLs registers extra CSS URLs for the shell.
func SetExtraStylesheetURLs(urls []string) {
	render.SetExtraStylesheetURLs(urls)
}
