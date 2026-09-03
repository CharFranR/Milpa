package integration

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"milpa/infrastructure/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	TestPool     *pgxpool.Pool
	TestContainer *postgres.PostgresContainer
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	dbCredentials := struct {
		dbName         string
		dbUser         string
		dbPassword     string
		migrationsPath string
	}{
		dbName:         "users",
		dbUser:         "user",
		dbPassword:     "password",
		migrationsPath: "../../infrastructure/adapters/secondary/repository/migrations",
	}

	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithOrderedInitScripts(
			filepath.Join(dbCredentials.migrationsPath, "000001_create_addresses.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000002_create_categories.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000003_create_users.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000004_create_companies.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000005_create_company_categories.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000006_create_offerings.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000007_create_inquiries.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000008_create_reviews.up.sql"),
		),
		postgres.WithDatabase(dbCredentials.dbName),
		postgres.WithUsername(dbCredentials.dbUser),
		postgres.WithPassword(dbCredentials.dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("failed to start container: %s", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := database.CreatePool(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}

	TestPool = pool
	TestContainer = postgresContainer

	code := m.Run()

	pool.Close()
	if err := postgresContainer.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %s", err)
	}

	os.Exit(code)
}
