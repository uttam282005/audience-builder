package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB connection for audience data.
type DB struct {
	*sql.DB
}

// Open initializes and migrates a SQLite database at the given path.
// Use ":memory:" for an in-memory database.
func Open(dataSourceName string) (*DB, error) {
	db, err := sql.Open("sqlite", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Recommended connection pooling settings for SQLite
	db.SetMaxOpenConns(1) // Avoid lock contention on writes

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	wrapped := &DB{DB: db}
	if err := wrapped.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return wrapped, nil
}

func (db *DB) migrate() error {
	schema := `
	PRAGMA foreign_keys = ON;

	CREATE TABLE IF NOT EXISTS users (
		anonymous_id TEXT PRIMARY KEY,
		created_at TEXT NOT NULL
	);

	CREATE TABLE IF NOT EXISTS events (
		id TEXT PRIMARY KEY,
		anonymous_id TEXT NOT NULL,
		event_type TEXT NOT NULL,
		timestamp TEXT NOT NULL,
		FOREIGN KEY(anonymous_id) REFERENCES users(anonymous_id)
	);

	CREATE INDEX IF NOT EXISTS idx_events_type_time ON events(event_type, timestamp, anonymous_id);
	CREATE INDEX IF NOT EXISTS idx_events_user ON events(anonymous_id);
	`
	_, err := db.Exec(schema)
	return err
}
