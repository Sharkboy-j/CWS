package store

import (
	"context"
	"cws/logger"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func ensureDatabase(host string, port int, user, password, dbname string) error {
	logger.Debug("Проверка существования БД %s", dbname)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Ошибка при подключении к postgres: %v", err)

		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	var exists bool
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbname,
	).Scan(&exists)
	if err != nil {
		logger.Error("Ошибка при проверке существования БД %s: %v", dbname, err)

		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		logger.Info("БД %s не существует, создание...", dbname)
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
		if err != nil {
			logger.Error("Ошибка при создании БД %s: %v", dbname, err)

			return fmt.Errorf("failed to create database: %w", err)
		}
		logger.Info("БД %s успешно создана", dbname)
	} else {
		logger.Debug("БД %s уже существует", dbname)
	}

	return nil
}

func (r *Repository) runMigrations(ctx context.Context) error {
	if err := r.createMigrationsTable(ctx); err != nil {

		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	appliedMigrations, err := r.getAppliedMigrations(ctx)
	if err != nil {

		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrationsDir := "store/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		migrationsDir = filepath.Join(wd, "store", "migrations")
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {

		return fmt.Errorf("failed to read migrations directory %s: %w", migrationsDir, err)
	}

	var migrationFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, entry.Name()))
		}
	}

	sort.Slice(migrationFiles, func(i, j int) bool {
		numI := extractMigrationNumber(migrationFiles[i])
		numJ := extractMigrationNumber(migrationFiles[j])

		return numI < numJ
	})

	for _, file := range migrationFiles {
		migrationName := filepath.Base(file)
		migrationNum := extractMigrationNumber(file)

		if appliedMigrations[migrationName] {

			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {

			return fmt.Errorf("failed to read migration %s: %w", migrationName, err)
		}

		tx, err := r.db.BeginTx(ctx, nil)
		if err != nil {

			return fmt.Errorf("failed to begin transaction for migration %s: %w", migrationName, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		if _, err := tx.ExecContext(ctx,
			"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW())",
			migrationName); err != nil {
			_ = tx.Rollback()

			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		if err := tx.Commit(); err != nil {
			logger.Error("Ошибка при коммите миграции %s: %v", migrationName, err)

			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}

		logger.Info("Применена миграция: %s (номер: %d)", migrationName, migrationNum)
	}

	return nil
}

func (r *Repository) createMigrationsTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := r.db.ExecContext(ctx, query)

	return err
}

func (r *Repository) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {

		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {

			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func extractMigrationNumber(filename string) int {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	if len(parts) == 0 {

		return 0
	}

	num, err := strconv.Atoi(parts[0])
	if err != nil {

		return 0
	}

	return num
}
