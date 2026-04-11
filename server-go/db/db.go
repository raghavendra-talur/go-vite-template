package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Connect(dbPath string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	var err error
	// Enable WAL mode and foreign keys via pragmas in the connection string
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(wal)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbPath)
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Database connected: %s", dbPath)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

// Migrate runs all pending migrations in order. Each migration is applied
// inside a transaction, and the schema_version is bumped atomically.
//
// To add a new migration, append a string to the `migrations` slice below.
// The index (0-based) is the migration number. Migrations that have already
// been applied (version <= current) are skipped.
func Migrate() error {
	// Bootstrap: create the version-tracking table if it doesn't exist.
	if _, err := DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER NOT NULL
	)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Read current version (0 if table is empty → no migrations applied yet).
	var current int
	row := DB.QueryRow(`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	if err := row.Scan(&current); err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}

	for i, m := range migrations {
		ver := i + 1 // migrations are 1-indexed
		if ver <= current {
			continue
		}
		tx, err := DB.Begin()
		if err != nil {
			return fmt.Errorf("migration %d: begin: %w", ver, err)
		}
		if _, err := tx.Exec(m); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d failed: %w", ver, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, ver); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: record version: %w", ver, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migration %d: commit: %w", ver, err)
		}
		log.Printf("Applied migration %d", ver)
	}

	log.Printf("Database at schema version %d (of %d)", max(current, len(migrations)), len(migrations))
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SchemaVersion returns the number of migrations defined.
func SchemaVersion() int {
	return len(migrations)
}

// ── Migrations ───────────────────────────────────────────────────────────────
// Append-only. Never edit or reorder existing entries.

var migrations = []string{
	// 1: API tokens for bearer auth
	`
	CREATE TABLE IF NOT EXISTS api_tokens (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		name TEXT NOT NULL,
		token TEXT NOT NULL UNIQUE,
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		last_used_at TEXT
	);
	`,

	// 2: Example items table — replace with your domain schema
	`
	CREATE TABLE IF NOT EXISTS items (
		id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
		title TEXT NOT NULL DEFAULT '',
		body TEXT NOT NULL DEFAULT '',
		created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
		updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
	);
	`,
}
