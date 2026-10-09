package database

import (
	"context"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"

	"milpa/infrastructure/adapters/secondary/repository"
)

func NewMigrationSource() (source.Driver, error) {
	return iofs.New(repository.Migrations, "migrations")
}

func MakeMigrations(ctx context.Context, DatabaseURL string) {
	src, err := NewMigrationSource()
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
}
