package applog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"sumeru/core/server/config"
)

var root *zap.Logger

func parseLogTimezone(s string) (*time.Location, string) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "local") {
		return time.Local, "Local"
	}
	if strings.EqualFold(s, "utc") {
		return time.UTC, "UTC"
	}
	loc, err := time.LoadLocation(s)
	if err != nil {
		return time.UTC, "UTC"
	}
	return loc, s
}

func newJSONEncoder() zapcore.Encoder {
	encCfg := zap.NewProductionEncoderConfig()
	loc := effectiveLocation()
	encCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format(time.RFC3339Nano))
	}
	return zapcore.NewJSONEncoder(encCfg)
}

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

	logPath := strings.TrimSpace(c.LogFile)
	if c.LogRolling && logPath == "" {
		return fmt.Errorf("log_rolling=true requires log_file")
	}

	level := zapcore.InfoLevel
	if c.DevMode {
		level = zapcore.DebugLevel
	}

	maxSize := c.LogMaxSizeMB
	if maxSize <= 0 {
		maxSize = 100
	}
	maxBackups := c.LogMaxBackups
	if maxBackups < 0 {
		maxBackups = 0
	}
	maxAge := c.LogMaxAgeDays
	if maxAge < 0 {
		maxAge = 0
	}

	var cores []zapcore.Core

	if c.LogStdout {
		cores = append(cores, zapcore.NewCore(newJSONEncoder(), zapcore.Lock(os.Stdout), level))
	}

	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			return fmt.Errorf("mkdir log dir: %w", err)
		}
		var sync zapcore.WriteSyncer
		if c.LogRolling {
			sync = zapcore.AddSync(&lumberjack.Logger{
				Filename:   logPath,
				MaxSize:    maxSize,
				MaxBackups: maxBackups,
				MaxAge:     maxAge,
				LocalTime:  true,
			})
		} else {
			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return fmt.Errorf("open log file %q: %w", logPath, err)
			}
			sync = zapcore.AddSync(f)
		}
		cores = append(cores, zapcore.NewCore(newJSONEncoder(), sync, level))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(newJSONEncoder(), zapcore.Lock(os.Stderr), level))
	}

	core := zapcore.NewTee(cores...)
	root = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	zap.ReplaceGlobals(root)

	zap.RedirectStdLog(root.Named("stdlib"))

	root.Sugar().Infow("logging initialized",
		"log_enabled", c.LogEnabled,
		"log_timezone", logTzName,
		"log_stdout", c.LogStdout,
		"log_rolling", c.LogRolling,
		"log_file", logPath,
		"log_max_size_mb", maxSize,
		"log_max_backups", maxBackups,
		"log_max_age_days", maxAge,
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
