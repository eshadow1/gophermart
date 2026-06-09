package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/eshadow1/gophermart/internal/configs"
	"github.com/eshadow1/gophermart/internal/loggers"
	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	defaultDriver             = "pgx5"
	defaultMaxOpenConnections = 20
	defaultMinOpenConnections = 5
	defaultConnMaxLifetime    = 10 * time.Minute
	defaultMaxConnIdleTime    = 5 * time.Minute
)

type builderConnect struct {
}

// NewBuilderConnect создает и возвращает новый экземпляр строителя пула соединений.
func NewBuilderConnect() *builderConnect {
	return &builderConnect{}
}

// CreateNewPool создает пул соединений с PostgreSQL, применяет миграции базы данных
// и проверяет доступность БД.
func (*builderConnect) CreateNewPool(cfg *configs.StorageConfig) (*pgxpool.Pool, error) {
	config, errParseConfig := pgxpool.ParseConfig(cfg.PathDB)
	if errParseConfig != nil {
		return nil, fmt.Errorf("error parse config: %w", errParseConfig)
	}

	config.MaxConns = defaultMaxOpenConnections
	config.MinConns = defaultMinOpenConnections
	config.MaxConnIdleTime = defaultMaxConnIdleTime
	config.MaxConnLifetime = defaultConnMaxLifetime

	pool, errPool := pgxpool.NewWithConfig(context.Background(), config)
	if errPool != nil {
		return nil, fmt.Errorf("error parse config: %w", errPool)
	}

	if errMigrate := runMigrationsWithDB(pool, "file://"+cfg.PathMigrations); errMigrate != nil {
		return nil, fmt.Errorf("error migrate: %w", errMigrate)
	}
	loggers.Log.Info("Migrate successful")

	checkContext, cancel := context.WithTimeout(context.Background(), defaultConnMaxLifetime)
	defer cancel()
	if errPing := pool.Ping(checkContext); errPing != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", errPing)
	}

	return pool, nil
}

// runMigrationsWithDB применяет миграции базы данных из указанного каталога.
func runMigrationsWithDB(db *pgxpool.Pool, migrationsPath string) error {
	sqlDB := stdlib.OpenDBFromPool(db)
	dbDriver, errInstance := pgxmigrate.WithInstance(sqlDB, &pgxmigrate.Config{})
	if errInstance != nil {
		return errInstance
	}

	m, errDBInstance := migrate.NewWithDatabaseInstance(
		migrationsPath,
		defaultDriver,
		dbDriver,
	)
	if errDBInstance != nil {
		return fmt.Errorf("failed to init migrate: %w", errDBInstance)
	}

	if errUpMigrate := m.Up(); errUpMigrate != nil && !errors.Is(errUpMigrate, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", errUpMigrate)
	}

	return nil
}
