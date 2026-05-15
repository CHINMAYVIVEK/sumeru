package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DbHost             string
	DbPort             string
	DbUser             string
	DbPass             string
	DbName             string
	DbSslMode          string
	HttpPort           string
	AddonsPath         string   // raw from file: comma-separated addon directory roots (see AddonPaths after AbsPaths)
	AddonPaths         []string // absolute: core/base (platform addons), then addons_path; filled by AbsPaths()
	SumeruHome         string   // optional: directory of standard sumeru repo (go.mod); used for core/base and default assets/templates if set
	AssetsPath         string   // static files (CSS/JS); default core/engine/assets
	TemplatesPath      string   // HTML templates; default core/engine/templates
	BrandCSS           string   // optional path to extra CSS (served as /static/brand.css)
	LogoPath           string   // optional image path (served as /static/app-logo)
	CompanyDisplayName string   // optional header label; else first core.company when module installed
	UserDisplayName    string   // optional header label; else first core.user when module installed
	LogFile            string   // optional log file path (absolutized in AbsPaths); see log_stdout / log_rolling
	LogStdout          bool     // when true, emit JSON logs to stdout (typical for Kubernetes)
	LogRolling         bool     // when true and log_file set, use size-based rotation (lumberjack); false for append-only or external rotation
	LogMaxSizeMB       int      // max megabytes per log file before rotation (default 100 when log_rolling)
	LogMaxBackups      int      // retained rotated files (0 = lumberjack default)
	LogMaxAgeDays      int      // delete rolled files older than N days (0 = no age limit)
	LogEnabled         bool     // log_enabled: when false, no Zap sinks and L(ctx) is no-op; stdlib log discarded
	LogTimezone        string   // log_timezone: UTC, Local (default), or IANA (e.g. Asia/Kolkata) for timestamps
	DevMode            bool     // dev_mode INI key; parseBoolKey(..., false) — debug Zap level and dev-only server paths
}

var AppConfig Config

// ConfigFileDir is the absolute directory of the last successfully loaded INI (set by LoadConfig).
// Relative addons_path segments resolve against this directory.
var ConfigFileDir string

func LoadConfig(path string) error {
	AppConfig = Config{
		DbSslMode:    defaultDbSSLMode,
		LogStdout:    true,
		LogEnabled:   true,
		LogMaxSizeMB: 100,
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	ConfigFileDir = filepath.Dir(absPath)

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, iniSectionPrefix) || strings.HasPrefix(line, iniCommentSemi) || strings.HasPrefix(line, iniCommentHash) {
			continue
		}

		parts := strings.SplitN(line, iniSeparator, iniKeyValueParts)
		if len(parts) != iniKeyValueParts {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case keyDbHost:
			AppConfig.DbHost = val
		case keyDbPort:
			AppConfig.DbPort = val
		case keyDbUser:
			AppConfig.DbUser = val
		case keyDbPassword:
			AppConfig.DbPass = val
		case keyDbName:
			AppConfig.DbName = val
		case keyDbSSLMode:
			AppConfig.DbSslMode = val
		case keyHTTPPort:
			AppConfig.HttpPort = val
		case keyAddonsPath:
			AppConfig.AddonsPath = val
		case keySumeruHome:
			AppConfig.SumeruHome = val
		case keyAssetsPath:
			AppConfig.AssetsPath = val
		case keyTemplatesPath:
			AppConfig.TemplatesPath = val
		case keyBrandCSS:
			AppConfig.BrandCSS = val
		case keyLogoPath:
			AppConfig.LogoPath = val
		case keyCompanyDisplayName:
			AppConfig.CompanyDisplayName = val
		case keyUserDisplayName:
			AppConfig.UserDisplayName = val
		case keyLogFile:
			AppConfig.LogFile = val
		case keyLogStdout:
			AppConfig.LogStdout = parseBoolKey(val, true)
		case keyLogRolling:
			AppConfig.LogRolling = parseBoolKey(val, false)
		case keyLogMaxSizeMB:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxSizeMB = n
			}
		case keyLogMaxBackups:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxBackups = n
			}
		case keyLogMaxAgeDays:
			if n, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
				AppConfig.LogMaxAgeDays = n
			}
		case keyLogEnabled:
			AppConfig.LogEnabled = parseBoolKey(val, true)
		case keyLogTimezone:
			AppConfig.LogTimezone = val
		case keyDevMode:
			AppConfig.DevMode = parseBoolKey(val, false)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if err := validateRequired(&AppConfig, absPath); err != nil {
		return err
	}

	// Default assets/templates paths are applied in AbsPaths() so sumeru_home can anchor
	// them under the standard tree; do not set repo-relative defaults here (would resolve
	// from the INI directory and break workspace configs next to ../sumeru).

	return nil
}

// parseBoolKey parses INI booleans; empty string returns defaultVal.
func parseBoolKey(val string, defaultVal bool) bool {
	s := strings.TrimSpace(strings.ToLower(val))
	if s == "" {
		return defaultVal
	}
	if s == "0" || s == "false" || s == "no" || s == "off" {
		return false
	}
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func validateRequired(c *Config, path string) error {
	if c.DbHost == "" {
		return fmt.Errorf(errFmtDbHostRequired, path)
	}
	if c.DbPort == "" {
		return fmt.Errorf(errFmtDbPortRequired, path)
	}
	if c.DbUser == "" {
		return fmt.Errorf(errFmtDbUserRequired, path)
	}
	if c.DbPass == "" {
		return fmt.Errorf(errFmtDbPasswordRequired, path)
	}
	if c.DbName == "" {
		return fmt.Errorf(errFmtDbNameRequired, path)
	}
	if c.HttpPort == "" {
		return fmt.Errorf(errFmtHTTPPortRequired, path)
	}
	if c.AddonsPath == "" {
		return fmt.Errorf(errFmtAddonsPathRequired, path)
	}
	return nil
}
