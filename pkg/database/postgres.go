package database

import (
	"fmt"
	"gophemart/internal/config"
	"gophemart/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	logger.Info().
		Str("host", cfg.PostgresDatabase.Host).
		Str("port", cfg.PostgresDatabase.Port).
		Str("dbname", cfg.PostgresDatabase.DBName).
		Str("user", cfg.PostgresDatabase.User).
		Msg("Connecting to PostgreSQL database")

	maskedDSN := fmt.Sprintf("postgres://%s:****@%s:%s/%s?sslmode=%s",
		cfg.PostgresDatabase.User,
		cfg.PostgresDatabase.Host,
		cfg.PostgresDatabase.Port,
		cfg.PostgresDatabase.DBName,
		cfg.PostgresDatabase.SSLMode)

	logger.Debug().
		Str("dsn", maskedDSN).
		Msg("Database connection string")

	start := time.Now()
	db, err := gorm.Open(postgres.Open(cfg.PostgresDatabase.URI), &gorm.Config{
		TranslateError: true,
	})
	duration := time.Since(start)

	if err != nil {
		logger.Error().
			Err(err).
			Str("dsn", maskedDSN).
			Dur("duration_ms", duration).
			Msg("Failed to connect to PostgreSQL")
		return nil, fmt.Errorf("failed to connect using URI: %w", err)
	}

	logger.Info().
		Dur("duration_ms", duration).
		Msg("PostgreSQL connection established successfully")

	db, err = setupConnectionPool(db, cfg.PostgresDatabase)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupConnectionPool(db *gorm.DB, pgConfig config.PostgresDatabaseConfig) (*gorm.DB, error) {
	logger.Info().
		Int("max_open_conns", pgConfig.MaxOpenConns).
		Int("max_idle_conns", pgConfig.MaxIdleConns).
		Dur("max_lifetime", pgConfig.MaxLifetime).
		Msg("Configuring database connection pool")

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to get underlying SQL DB")
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(pgConfig.MaxOpenConns)
	sqlDB.SetMaxIdleConns(pgConfig.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(pgConfig.MaxLifetime)

	logger.Debug().Msg("Pinging database to verify connection")

	start := time.Now()
	if err := sqlDB.Ping(); err != nil {
		logger.Error().
			Err(err).
			Dur("duration_ms", time.Since(start)).
			Msg("Database ping failed")
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info().
		Dur("duration_ms", time.Since(start)).
		Msg("Database ping successful")

	logger.Info().
		Int("max_open_connections", sqlDB.Stats().MaxOpenConnections).
		Int("open_connections", sqlDB.Stats().OpenConnections).
		Int("in_use", sqlDB.Stats().InUse).
		Int("idle", sqlDB.Stats().Idle).
		Msg("Database connection pool initialized")
	return db, nil
}
