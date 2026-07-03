// Package config handles loading, validation, and access to AgentNetra's
// runtime configuration. Configuration is sourced from (in priority order):
//  1. Command-line flags
//  2. Environment variables (prefix: AGENTNETRA_)
//  3. Config file (YAML/TOML/JSON)
//  4. Built-in defaults
//
// This package wraps viper to provide a strongly-typed Config struct so that
// the rest of the application never touches viper directly.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"

	"github.com/chethanyadav456/agentnetra/pkg/constants"
)

// Config is the application-wide configuration object.
// All fields have sane defaults and can be overridden by the user.
type Config struct {
	// Log contains logging-related settings.
	Log LogConfig `mapstructure:"log"`

	// Database contains storage-related settings.
	Database DatabaseConfig `mapstructure:"database"`

	// Scanner contains process scanning settings.
	Scanner ScannerConfig `mapstructure:"scanner"`
}

// LogConfig holds settings that control the logger behaviour.
type LogConfig struct {
	// Level is the minimum log level. Accepted: debug, info, warn, error.
	Level string `mapstructure:"level"`

	// Development enables verbose, human-friendly log output.
	Development bool `mapstructure:"development"`

	// File is the optional path to a log file. Empty means stderr only.
	File string `mapstructure:"file"`
}

// DatabaseConfig holds settings for the SQLite storage layer.
type DatabaseConfig struct {
	// Path is the absolute path to the SQLite database file.
	Path string `mapstructure:"path"`
}

// ScannerConfig holds settings that govern process scanning behaviour.
type ScannerConfig struct {
	// IntervalSeconds is the pause between automatic scans (0 = one-shot).
	IntervalSeconds int `mapstructure:"interval_seconds"`
}

// Load reads configuration from the provided config file path (may be empty),
// environment variables, and built-in defaults. It returns a fully populated
// Config or a descriptive error.
func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	// --- Defaults ---
	setDefaults(v)

	// --- Environment variables ---
	v.SetEnvPrefix("AGENTNETRA")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// --- Config file ---
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(filepath.Join(home, ".agentnetra"))
			v.AddConfigPath(".")
		}
		v.SetConfigName("config")
		v.SetConfigType("yaml")
	}

	if err := v.ReadInConfig(); err != nil {
		// It is fine if the config file does not exist — use defaults.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("config: read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal config: %w", err)
	}

	return &cfg, nil
}

// setDefaults registers safe default values into the viper instance.
func setDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.development", false)
	v.SetDefault("log.file", defaultLogPath())
	v.SetDefault("database.path", defaultDBPath())
	v.SetDefault("scanner.interval_seconds", 0)
}

// defaultDBPath returns the platform-appropriate path for the SQLite database.
func defaultDBPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Application Support", "AgentNetra", constants.DBFileName)
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "AgentNetra", constants.DBFileName)
	default: // Linux and other Unix-like systems.
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".agentnetra", constants.DBFileName)
	}
}

// defaultLogPath returns the platform-appropriate directory for log files.
func defaultLogPath() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", "AgentNetra", "agentnetra.log")
	case "windows":
		appData := os.Getenv("APPDATA")
		return filepath.Join(appData, "AgentNetra", "logs", "agentnetra.log")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".agentnetra", "logs", "agentnetra.log")
	}
}
