package integration

import (
	"context"
	"log"
	"path/filepath"

	"milpa/infrastructure/database"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func CreateTestContainer() (*postgres.PostgresContainer, error) {

	db_credentials := struct {
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

	postgresContainer, err := postgres.Run(context.Background(),
		"postgres:16-alpine",
		postgres.WithOrderedInitScripts(
			// Deuda tecnica, hardcode
			filepath.Join(db_credentials.migrationsPath, "000001_create_addresses.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000002_create_categories.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000003_create_users.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000004_create_companies.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000005_create_company_categories.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000006_create_offerings.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000007_create_inquiries.up.sql"),
			filepath.Join(db_credentials.migrationsPath, "000008_create_reviews.up.sql"),
		),
		postgres.WithDatabase(db_credentials.dbName),
		postgres.WithUsername(db_credentials.dbUser),
		postgres.WithPassword(db_credentials.dbPassword),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {

		log.Printf("failed to start container: %s", err)
		return nil, err
	}

	return postgresContainer, err
}

func InitTestDB() (*pgxpool.Pool, *postgres.PostgresContainer, error) {
	db, err := CreateTestContainer()

	if err != nil {
		log.Printf("failed no init test DB: %v", err)

	}

	DatabaseURL, err := db.ConnectionString(context.Background())

	PoolConnection, err := database.CreatePool(context.Background(), DatabaseURL)

	return PoolConnection, db, err
}
