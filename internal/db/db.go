package db

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	*sql.DB
}

func Open(ctx context.Context, dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0750); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode and other optimizations
	if _, err := db.ExecContext(ctx, `
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA busy_timeout = 5000;
		PRAGMA foreign_keys = ON;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure sqlite: %w", err)
	}

	d := &DB{DB: db}
	if err := d.initSchema(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return d, nil
}

func (d *DB) initSchema(ctx context.Context) error {
	// Check if tables exist
	var count int
	err := d.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM sqlite_master 
		WHERE type='table' AND name IN ('admin_users', 'persons', 'files')
	`).Scan(&count)
	if err != nil {
		return err
	}

	if count >= 3 {
		// Tables already exist, try to migrate if needed
		return d.migrate(ctx)
	}

	// Create initial schema
	if _, err := d.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	return nil
}

func (d *DB) migrate(ctx context.Context) error {
	// Add any migration logic here
	// For now, just ensure all expected columns exist
	return nil
}

func (d *DB) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

// Backup creates a backup of the database using SQLite backup API
func (d *DB) Backup(ctx context.Context, backupPath string) error {
	if err := os.MkdirAll(filepath.Dir(backupPath), 0750); err != nil {
		return fmt.Errorf("create backup dir: %w", err)
	}

	// Use VACUUM INTO for safe online backup
	_, err := d.ExecContext(ctx, `VACUUM INTO ?`, backupPath)
	if err != nil {
		// Fallback to simple copy if VACUUM INTO not supported
		src, err := os.Open(d.DB.Driver().(*sqlite.Driver).DataSourceName())
		if err != nil {
			return fmt.Errorf("open source: %w", err)
		}
		defer src.Close()

		dst, err := os.Create(backupPath)
		if err != nil {
			return fmt.Errorf("create dest: %w", err)
		}
		defer dst.Close()

		if _, err := dst.ReadFrom(src); err != nil {
			return fmt.Errorf("copy: %w", err)
		}
	}

	return nil
}

// WithTransaction executes a function within a transaction
func (d *DB) WithTransaction(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := d.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}
