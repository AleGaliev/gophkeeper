package migrations

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const migrationsPath = "file://internal/infra/postgres/migrations/file"

type logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Debug(ctx context.Context, msg string, args ...any)
}

type Migration struct {
	migration *migrate.Migrate
	logger
}

func New(migrationDSN string, logger logger) (*Migration, error) {
	m, err := migrate.New(migrationsPath, migrationDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %v", err)
	}
	return &Migration{
		migration: m,
		logger:    logger,
	}, nil
}

func (m *Migration) Up() error {
	if err := m.migration.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migration: %v", err)
	}
	m.logger.Info(context.Background(), "migrate up successfully")
	return nil
}

func (m *Migration) Down() error {
	if err := m.migration.Down(); err != nil {
		return fmt.Errorf("failed to run migration: %v", err)
	}
	m.logger.Info(context.Background(), "migrate down successfully")
	return nil
}
