package database

import (
	"fmt"
	"gophemart/internal/app/entity"
	"gophemart/pkg/logger"
	"gorm.io/gorm"
	"time"
)

func Migrate(db *gorm.DB) error {
	models := []struct {
		name  string
		model interface{}
	}{
		{"User", &entity.User{}},
		{"Order", &entity.Order{}},
		{"Withdrawal", &entity.Withdrawal{}},
	}

	logger.Info().Msg("Starting database migration")
	logger.Info().Int("model_count", len(models)).Msg("Models to migrate")

	for _, m := range models {
		start := time.Now()
		logger.Info().
			Str("model", m.name).
			Msg("Migrating model")

		if err := db.AutoMigrate(m.model); err != nil {
			logger.Error().
				Err(err).
				Str("model", m.name).
				Msg("Migration failed for model")
			return fmt.Errorf("failed to migrate %s: %w", m.name, err)
		}

		duration := time.Since(start)
		logger.Info().
			Str("model", m.name).
			Dur("duration_ms", duration).
			Msg("Model migrated successfully")
	}

	logger.Info().Msg("Database migration completed successfully")
	return nil
}
