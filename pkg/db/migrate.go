package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunDuckDBMigrations applies migrations from a directory to a DuckDB database.
func RunDuckDBMigrations(db *sql.DB, migrationsDir string) error {
	// 1. Ensure the schema_migrations table exists.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// 2. Get the list of applied migrations.
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating over applied migrations: %w", err)
	}

	// 3. Read migration files from the directory.
	migrationFiles, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort files to apply them in order.
	sort.Slice(migrationFiles, func(i, j int) bool {
		return migrationFiles[i].Name() < migrationFiles[j].Name()
	})

	// 4. Apply pending migrations.
	for _, file := range migrationFiles {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		version := strings.Split(file.Name(), "_")[0]
		if appliedMigrations[version] {
			continue
		}

		// Apply the migration
		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", file.Name(), err)
		}

		if _, err := tx.Exec(string(content)); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %s: %w", file.Name(), err)
		}

		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", file.Name(), err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for migration %s: %w", file.Name(), err)
		}
	}

	return nil
}
