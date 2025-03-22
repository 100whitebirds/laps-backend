package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type MigrationRecord struct {
	Version   string
	Name      string
	AppliedAt time.Time
}

func RunMigrations(db *pgxpool.Pool, migrationsDir string, logger *zap.Logger) error {
	ctx := context.Background()

	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			version VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("ошибка при создании таблицы миграций: %w", err)
	}

	var appliedMigrations []MigrationRecord
	rows, err := db.Query(ctx, "SELECT version, name, applied_at FROM migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("ошибка при получении списка выполненных миграций: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var record MigrationRecord
		if err := rows.Scan(&record.Version, &record.Name, &record.AppliedAt); err != nil {
			return fmt.Errorf("ошибка при сканировании записи о миграции: %w", err)
		}
		appliedMigrations = append(appliedMigrations, record)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("ошибка при обработке результатов запроса: %w", err)
	}

	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("ошибка при чтении директории миграций: %w", err)
	}

	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	appliedMap := make(map[string]bool)
	for _, migration := range appliedMigrations {
		appliedMap[migration.Version] = true
	}

	for _, file := range migrationFiles {
		parts := strings.SplitN(file, "_", 2)
		if len(parts) != 2 {
			logger.Warn("неверный формат имени файла миграции", zap.String("file", file))
			continue
		}

		version := parts[0]
		name := strings.TrimSuffix(parts[1], ".sql")

		if appliedMap[version] {
			logger.Info("миграция уже выполнена", zap.String("version", version), zap.String("name", name))
			continue
		}

		filePath := filepath.Join(migrationsDir, file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("ошибка при чтении файла миграции %s: %w", file, err)
		}

		logger.Info("выполнение миграции", zap.String("version", version), zap.String("name", name))

		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("ошибка при начале транзакции: %w", err)
		}

		_, err = tx.Exec(ctx, string(content))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("ошибка при выполнении миграции %s: %w", file, err)
		}

		_, err = tx.Exec(ctx,
			"INSERT INTO migrations (version, name, applied_at) VALUES ($1, $2, $3)",
			version, name, time.Now(),
		)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("ошибка при записи информации о выполненной миграции: %w", err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("ошибка при коммите транзакции: %w", err)
		}

		logger.Info("миграция выполнена успешно", zap.String("version", version), zap.String("name", name))
	}

	return nil
}
