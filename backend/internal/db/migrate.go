package db

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(db *sql.DB) error {
	ctx := context.Background()

	if err := createMigrationsTable(ctx, db); err != nil {
		return err
	}

	applied, err := appliedMigrations(ctx, db)
	if err != nil {
		return err
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migration directory: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		if applied[name] {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", name, err)
		}

		if err := applyMigration(ctx, db, name, string(content)); err != nil {
			return fmt.Errorf("applying migration %s: %w", name, err)
		}
	}

	return nil
}

func createMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}
	return nil
}

func appliedMigrations(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("querying applied migrations: %w", err)
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning migration name: %w", err)
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, db *sql.DB, name, content string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, content); err != nil {
		return fmt.Errorf("executing SQL: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name) VALUES (?)", name); err != nil {
		return fmt.Errorf("recording migration: %w", err)
	}

	return tx.Commit()
}
