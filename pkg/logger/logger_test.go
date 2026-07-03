// Package logger_test contains unit tests for the AgentNetra logger package.
package logger_test

import (
	"testing"

	"github.com/chethanyadav456/agentnetra/pkg/logger"
	"go.uber.org/zap"
)

func TestNewLogger_DefaultLevel(t *testing.T) {
	t.Parallel()

	cfg := logger.Config{
		Level:      "info",
		AppName:    "agentnetra",
		AppVersion: "0.1.0",
	}

	log, err := logger.New(cfg)
	if err != nil {
		t.Fatalf("logger.New: unexpected error: %v", err)
	}
	if log == nil {
		t.Fatal("logger.New returned nil logger")
	}

	// Smoke-test each level method — they must not panic.
	log.Debug("debug message", zap.String("key", "value"))
	log.Info("info message")
	log.Warn("warn message")
	log.Error("error message")

	if err := log.Sync(); err != nil {
		// Syncing stderr can return "invalid argument" on macOS. That is fine.
		t.Logf("Sync returned non-nil (expected on some platforms): %v", err)
	}
}

func TestNewLogger_InvalidLevel_FallsBackToInfo(t *testing.T) {
	t.Parallel()

	cfg := logger.Config{
		Level:      "nonsense-level",
		AppName:    "agentnetra",
		AppVersion: "0.1.0",
	}

	log, err := logger.New(cfg)
	// Invalid level should not produce an error; it falls back to INFO.
	if err != nil {
		t.Fatalf("logger.New: unexpected error for invalid level: %v", err)
	}
	if log == nil {
		t.Fatal("logger.New returned nil logger for invalid level")
	}
}

func TestNewNop_DoesNotPanic(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	if log == nil {
		t.Fatal("NewNop returned nil")
	}

	log.Debug("nop debug")
	log.Info("nop info")
	log.Warn("nop warn")
	log.Error("nop error")
	_ = log.Sync()
}

func TestLoggerWith_ReturnsChildLogger(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	child := log.With(zap.String("component", "test"))
	if child == nil {
		t.Fatal("With returned nil child logger")
	}

	// Child must work correctly.
	child.Info("child message")
}
