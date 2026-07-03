// Package storage provides the SQLite persistence layer for AgentNetra.
// It exposes a DB type that manages the database connection and runs
// migrations, plus repository types (AgentRepository, SessionRepository,
// EventRepository) that implement the data access layer.
//
// All database operations use the pure-Go modernc.org/sqlite driver so no
// CGo is required and cross-compilation works out of the box.
package storage

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite" // Register the "sqlite" driver.

	"go.uber.org/zap"

	"github.com/chethanyadav456/agentnetra/pkg/logger"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps the underlying sql.DB connection and provides the migration runner.
// Always obtain a DB via Open(); never construct directly.
type DB struct {
	conn *sql.DB
	log  *logger.Logger
}

// Open opens (or creates) the SQLite database at the given path, applies any
// pending migrations, and returns a ready-to-use DB. The caller is responsible
// for calling Close() when finished.
func Open(path string, log *logger.Logger) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return nil, fmt.Errorf("storage: create database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("storage: open database: %w", err)
	}

	// SQLite works best with a single writer connection.
	conn.SetMaxOpenConns(1)

	// Enable WAL mode and foreign key enforcement.
	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, p := range pragmas {
		if _, err := conn.Exec(p); err != nil {
			_ = conn.Close()
			return nil, fmt.Errorf("storage: set pragma (%s): %w", p, err)
		}
	}

	db := &DB{conn: conn, log: log}

	if err := db.migrate(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("storage: run migrations: %w", err)
	}

	return db, nil
}

// Close releases the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn exposes the raw *sql.DB for use by repository types.
// Only repository types in this package should call Conn().
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// migrate applies all SQL migration files embedded in the migrations/ directory
// that have not yet been recorded in the schema_migrations table. Migration
// files are applied in lexicographic (filename) order, which aligns with the
// numeric prefixes used in the filenames (001_, 002_, …).
func (db *DB) migrate() error {
	// Bootstrap the tracking table first so we can record applied migrations.
	bootstrap := `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT     PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);`
	if _, err := db.conn.Exec(bootstrap); err != nil {
		return fmt.Errorf("migrate: bootstrap schema_migrations: %w", err)
	}

	// Discover all embedded *.sql files.
	entries, err := fs.Glob(migrationsFS, "migrations/*.sql")
	if err != nil {
		return fmt.Errorf("migrate: list migration files: %w", err)
	}
	sort.Strings(entries)

	for _, entry := range entries {
		version := filepath.Base(entry)

		// Skip the schema_migrations bootstrap file to avoid circular dependency.
		if strings.Contains(version, "schema_migrations") {
			continue
		}

		// Check whether this version has already been applied.
		var count int
		err := db.conn.QueryRow(
			`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, version,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("migrate: check %s: %w", version, err)
		}
		if count > 0 {
			db.log.Debug("storage: migration already applied", zap.String("version", version))
			continue
		}

		// Read and execute the migration SQL.
		content, err := migrationsFS.ReadFile(entry)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", version, err)
		}

		if _, err := db.conn.Exec(string(content)); err != nil {
			return fmt.Errorf("migrate: execute %s: %w", version, err)
		}

		// Record the applied migration.
		if _, err := db.conn.Exec(
			`INSERT INTO schema_migrations (version) VALUES (?)`, version,
		); err != nil {
			return fmt.Errorf("migrate: record %s: %w", version, err)
		}

		db.log.Info("storage: migration applied", zap.String("version", version))
	}

	return nil
}
