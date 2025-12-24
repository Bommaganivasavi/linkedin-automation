// internal/storage/storage.go
package storage

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // This is the pure Go SQLite driver
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("failed to enable foreign keys: %v", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %v", err)
	}

	return &Storage{db: db}, nil
}

func createTables(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS connection_requests (
			id TEXT PRIMARY KEY,
			profile_url TEXT NOT NULL UNIQUE,
			profile_name TEXT NOT NULL,
			sent_at TEXT NOT NULL,
			accepted BOOLEAN NOT NULL DEFAULT 0,
			note TEXT,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}
