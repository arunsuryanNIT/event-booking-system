package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// LogFunc is a callback for logging migration progress.
// This avoids coupling the db package to a specific logger implementation.
type LogFunc func(msg string, kvs ...interface{})

// RunMigrations reads all .sql files from migrationsDir in lexical order
// and executes each one against db. Migrations are expected to be idempotent
// (using IF NOT EXISTS, ON CONFLICT DO NOTHING, etc.) so re-running is safe.
// The logFn callback is invoked after each successful migration.
func RunMigrations(db *sql.DB, migrationsDir string, logFn LogFunc) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)

	for _, file := range files {
		path := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", file, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("executing migration %s: %w", file, err)
		}

		if logFn != nil {
			logFn("migration applied", "file", file)
		}
	}

	return nil
}
