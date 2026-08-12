package applog

import (
	"io"
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"sumeru/core/server/config"
)

var root *zap.Logger

// SetupFromConfig builds the global Zap logger and redirects the standard library log package.
// Call after config.LoadConfig and config.AbsPaths so log_file paths are absolute.
//
//   - log_enabled=false: no log sinks, stdlib log discarded, L(ctx) is a no-op (use for benchmarks / quiet CI).
//   - log_stdout=true (default): JSON logs to stdout (typical Kubernetes / containers).
//   - log_file: optional second sink; log_rolling=true uses lumberjack (VPS / single host).
//   - log_timezone: UTC, Local (default), or IANA name (e.g. Asia/Kolkata) for Zap EncodeTime and log_ts in L(ctx).
func SetupFromConfig(c *config.Config) error {
	if root != nil {
		_ = root.Sync()
		root = nil
	}

	logEnabled = c.LogEnabled
	loc, tzLabel := parseLogTimezone(c.LogTimezone)
	logLocation = loc
	logTzName = tzLabel

	if !c.LogEnabled {
		root = zap.NewNop()
		zap.ReplaceGlobals(root)
		log.SetOutput(io.Discard)
		return nil
	}
	log.SetOutput(os.Stderr)

	level := zapcore.InfoLevel
	if c.DevMode {
		level = zapcore.DebugLevel
	}

	sinks, err := buildCores(c, level)
	if err != nil {
		return err
	}

	core := zapcore.NewTee(sinks.cores...)
	root = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(root)

	zap.RedirectStdLog(root.Named("stdlib"))

	root.Sugar().Infow("logging initialized",
		"log_enabled", c.LogEnabled,
		"log_timezone", logTzName,
		"log_stdout", c.LogStdout,
		"log_rolling", c.LogRolling,
		"log_file", sinks.logPath,
		"log_max_size_mb", sinks.maxSize,
		"log_max_backups", sinks.maxBackups,
		"log_max_age_days", sinks.maxAge,
	)
	return nil
}

// Sync flushes buffered log entries (call on graceful shutdown).
func Sync() {
	if root != nil {
		_ = root.Sync()
	}
}

// Sugar returns the global sugared logger (nil if SetupFromConfig was not called).
func Sugar() *zap.SugaredLogger {
	if root == nil {
		return nil
	}
	return root.Sugar()
}
