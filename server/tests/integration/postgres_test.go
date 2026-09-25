package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"milpa/infrastructure/database"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	TestPool       *pgxpool.Pool
	TestContainer  *postgres.PostgresContainer
	TestESClient   *elasticsearch.Client
	TestESEndpoint string
	TestRedisAddr  string
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
			filepath.Join(dbCredentials.migrationsPath, "000009_create_liquidations.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000010_create_reports.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000011_create_audit_logs.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000012_add_user_suspended_at.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000013_create_conversations.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000014_create_messages.up.sql"),
			filepath.Join(dbCredentials.migrationsPath, "000015_add_conversation_message_indexes.up.sql"),
		),
		postgres.WithDatabase(dbCredentials.dbName),
		postgres.WithUsername(dbCredentials.dbUser),
		postgres.WithPassword(dbCredentials.dbPassword),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %s", err)
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

	// Elasticsearch container
	esContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "elasticsearch:8.12.0",
			ExposedPorts: []string{"9200/tcp"},
			Env: map[string]string{
				"discovery.type":                  "single-node",
				"xpack.security.enabled":          "false",
				"xpack.security.http.ssl.enabled": "false",
				"ES_JAVA_OPTS":                    "-Xms256m -Xmx256m",
			},
			WaitingFor: wait.ForHTTP("/_cluster/health").
				WithPort("9200").
				WithStartupTimeout(60 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to start elasticsearch container: %s", err)
	}

	esHost, _ := esContainer.Host(ctx)
	esPort, _ := esContainer.MappedPort(ctx, "9200")
	esEndpoint := fmt.Sprintf("http://%s:%s", esHost, esPort.Port())
	TestESEndpoint = esEndpoint

	esCfg := elasticsearch.Config{
		Addresses: []string{esEndpoint},
	}
	esClient, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Fatalf("failed to create ES client: %v", err)
	}
	TestESClient = esClient

	// Redis container
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "redis:7-alpine",
			ExposedPorts: []string{"6379/tcp"},
			WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(10 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		log.Fatalf("failed to start redis container: %s", err)
	}

	redisHost, _ := redisContainer.Host(ctx)
	redisPort, _ := redisContainer.MappedPort(ctx, "6379")
	TestRedisAddr = fmt.Sprintf("%s:%s", redisHost, redisPort.Port())

	code := m.Run()

	pool.Close()
	postgresContainer.Terminate(ctx)
	esContainer.Terminate(ctx)
	redisContainer.Terminate(ctx)

	os.Exit(code)
}

// cleanupTables truncates all tables in foreign-key-safe order so each test
// starts with a clean database. Call this at the top of every setup*TestData
// helper or directly in a Test function.
func cleanupTables(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	tables := []string{
		"audit_logs",
		"reports",
		"reviews",
		"inquiries",
		"liquidations",
		"offerings",
		"company_categories",
		"companies",
		"users",
		"categories",
		"addresses",
	}
	for _, tbl := range tables {
		_, err := TestPool.Exec(ctx, "TRUNCATE TABLE "+tbl+" CASCADE")
		if err != nil {
			t.Fatalf("cleanupTables: truncate %s: %v", tbl, err)
		}
	}
}
