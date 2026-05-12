package base

import "sumeru/core/server/config"

// Config is the on-disk configuration shape (INI [options]).
type Config = config.Config

// Cfg returns the live process configuration (same memory as the INI-backed singleton).
func Cfg() *Config {
	return &config.AppConfig
}
