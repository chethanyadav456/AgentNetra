// Package logger provides a structured, levelled logging facility for
// AgentNetra built on top of go.uber.org/zap. It exposes a thin wrapper
// that injects context fields (version, app name) into every log entry and
// offers a consistent interface throughout the application.
package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the AgentNetra application logger. It wraps zap.Logger and
// provides convenience methods with pre-attached context fields.
type Logger struct {
	zl *zap.Logger
}

// Config holds the configuration options used to initialise the logger.
type Config struct {
	// Level is the minimum log level to emit. Accepted values: debug, info, warn, error.
	Level string

	// Development enables full debug output and caller information when true.
	Development bool

	// LogFile is the optional path to a file where logs are written.
	// When empty, output goes to stderr only.
	LogFile string

	// AppName is injected as a static field into every log entry.
	AppName string

	// AppVersion is injected as a static field into every log entry.
	AppVersion string
}

// New creates and returns a configured Logger instance. If the provided
// Config is invalid, New returns a non-nil error describing the problem.
func New(cfg Config) (*Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		// Fall back to INFO rather than failing hard on an unrecognised level.
		level = zapcore.InfoLevel
	}

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Build cores — always write to stderr; optionally also to a file.
	cores := []zapcore.Core{
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(os.Stderr),
			level,
		),
	}

	if cfg.LogFile != "" {
		if err := os.MkdirAll(filepath.Dir(cfg.LogFile), 0o750); err != nil {
			return nil, fmt.Errorf("logger: create log directory: %w", err)
		}

		f, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
		if err != nil {
			return nil, fmt.Errorf("logger: open log file: %w", err)
		}

		cores = append(cores, zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderCfg),
			zapcore.AddSync(f),
			level,
		))
	}

	opts := []zap.Option{zap.AddCallerSkip(1)}
	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	zl := zap.New(zapcore.NewTee(cores...), opts...).With(
		zap.String("app", cfg.AppName),
		zap.String("version", cfg.AppVersion),
	)

	return &Logger{zl: zl}, nil
}

// NewNop returns a no-op Logger that discards all log output. Useful in tests
// and any context where logging is undesirable.
func NewNop() *Logger {
	return &Logger{zl: zap.NewNop()}
}

// Debug logs a message at DEBUG level with optional key-value fields.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.zl.Debug(msg, fields...)
}

// Info logs a message at INFO level with optional key-value fields.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.zl.Info(msg, fields...)
}

// Warn logs a message at WARN level with optional key-value fields.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.zl.Warn(msg, fields...)
}

// Error logs a message at ERROR level with optional key-value fields.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.zl.Error(msg, fields...)
}

// With returns a child Logger that automatically includes the supplied fields
// in every subsequent log entry.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{zl: l.zl.With(fields...)}
}

// Sync flushes any buffered log entries. Always call this before the process
// exits (typically deferred in main).
func (l *Logger) Sync() error {
	return l.zl.Sync()
}
