package server

import (
	"sumeru/core/base"
	"sumeru/core/server/config"
)

// Config is the on-disk configuration shape (stable alias via base).
type Config = base.Config

// Cfg returns a pointer to the live process configuration.
func Cfg() *config.Config {
	return base.Cfg()
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
	base.InitDB(base.InitDBInput{DSN: databaseConnectionString})
}

// SyncModels syncs registered models to the database.
func SyncModels() error {
	return base.SyncModels(base.SyncModelsInput{})
}

// LoadAddonPaths discovers and loads addons from the given roots.
func LoadAddonPaths(paths []string) error {
	return base.LoadAddonPaths(base.LoadAddonPathsInput{Paths: paths})
}

// RunModuleCLI runs -i / -u module lists.
func RunModuleCLI(installCSV, updateCSV string) error {
	return base.RunModuleCLI(base.RunModuleCLIInput{Install: installCSV, Update: updateCSV})
}

// SetExtraStylesheetURLs registers extra CSS URLs for the shell.
func SetExtraStylesheetURLs(urls []string) {
	base.SetExtraStylesheetURLs(base.SetExtraStylesheetURLsInput{URLs: urls})
}
