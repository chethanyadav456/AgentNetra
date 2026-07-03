// Package config_test contains unit tests for the AgentNetra config package.
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/chethanyadav456/agentnetra/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load: unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("config.Load returned nil config")
	}

	// Log level must have a non-empty default.
	if cfg.Log.Level == "" {
		t.Error("log.level default is empty; expected 'info'")
	}

	// Database path must be non-empty.
	if cfg.Database.Path == "" {
		t.Error("database.path default is empty")
	}
}

func TestLoad_FromFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")

	content := `
log:
  level: debug
  development: true
database:
  path: /tmp/test_agentnetra.db
scanner:
  interval_seconds: 30
`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("config.Load from file: unexpected error: %v", err)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("log.level: got %q, want %q", cfg.Log.Level, "debug")
	}
	if !cfg.Log.Development {
		t.Error("log.development: got false, want true")
	}
	if cfg.Database.Path != "/tmp/test_agentnetra.db" {
		t.Errorf("database.path: got %q, want %q", cfg.Database.Path, "/tmp/test_agentnetra.db")
	}
	if cfg.Scanner.IntervalSeconds != 30 {
		t.Errorf("scanner.interval_seconds: got %d, want 30", cfg.Scanner.IntervalSeconds)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	// Cannot run in parallel because it mutates environment.
	t.Setenv("AGENTNETRA_LOG_LEVEL", "warn")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load: unexpected error: %v", err)
	}

	if cfg.Log.Level != "warn" {
		t.Errorf("log.level: got %q, want %q", cfg.Log.Level, "warn")
	}
}

func TestLoad_NonexistentFile_ReturnsError(t *testing.T) {
	t.Parallel()

	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent config file, got nil")
	}
}
