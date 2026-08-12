package applog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"sumeru/core/server/config"
)

type sinkResult struct {
	cores      []zapcore.Core
	logPath    string
	maxSize    int
	maxBackups int
	maxAge     int
}

func buildCores(c *config.Config, level zapcore.Level) (sinkResult, error) {
	logPath := strings.TrimSpace(c.LogFile)
	if c.LogRolling && logPath == "" {
		return sinkResult{}, fmt.Errorf("log_rolling=true requires log_file")
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
		sync, err := fileSyncer(logPath, c.LogRolling, maxSize, maxBackups, maxAge)
		if err != nil {
			return sinkResult{}, err
		}
		cores = append(cores, zapcore.NewCore(newJSONEncoder(), sync, level))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(newJSONEncoder(), zapcore.Lock(os.Stderr), level))
	}

	return sinkResult{
		cores:      cores,
		logPath:    logPath,
		maxSize:    maxSize,
		maxBackups: maxBackups,
		maxAge:     maxAge,
	}, nil
}

func fileSyncer(path string, rolling bool, maxSize, maxBackups, maxAge int) (zapcore.WriteSyncer, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir log dir: %w", err)
	}
	if rolling {
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			LocalTime:  true,
		}), nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open log file %q: %w", path, err)
	}
	return zapcore.AddSync(f), nil
}
