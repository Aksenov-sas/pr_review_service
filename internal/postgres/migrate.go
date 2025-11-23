package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/pressly/goose/v3"
)

// Migrate выполняет миграции базы данных
func Migrate(db *sql.DB, logger *slog.Logger) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("ошибка установки диалекта: %w", err)
	}

	// Применяем все миграции из директории migrations
	possiblePaths := []string{"./migrations", "/app/migrations", "../migrations", "../../migrations"}

	for _, path := range possiblePaths {
		logger.Info("Попытка выполнения миграций", slog.String("path", path))
		err := goose.Up(db, path)
		if err == nil {
			logger.Info("Миграции успешно выполнены", slog.String("path", path))
			return nil
		} else {
			logger.Error("Ошибка выполнения миграций по пути", slog.String("path", path), slog.String("error", err.Error()))
		}
	}

	return fmt.Errorf("не удалось выполнить миграции ни по одному из путей")
}
