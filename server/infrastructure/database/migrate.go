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

// NewMigrationSource builds the source driver over the migrations compiled into
// the binary by the repository package.
//
// It is exported so the integration suite can assemble the same source the
// runner uses and prove it resolves with an arbitrary working directory, but
// it takes no arguments and touches nothing outside the process: the returned
// driver reads from an in-memory filesystem, so building it is side-effect free
// and safe to call as many times as a test needs.
func NewMigrationSource() (source.Driver, error) {
	return iofs.New(repository.Migrations, "migrations")
}

// MakeMigrations applies every pending migration to DatabaseURL.
//
// The source is the embedded filesystem, never a path: resolving
// "file://infrastructure/..." against the process working directory meant the
// server only booted when it was launched from the module root, which is a
// property of the invocation rather than of the binary.
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
