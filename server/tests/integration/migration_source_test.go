package integration

import (
	"context"
	"errors"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"

	"milpa/infrastructure/adapters/secondary/repository"
	"milpa/infrastructure/database"
)

// migrationsDirOnDisk is the same directory the TestContainers harness replays
// from, which makes it the reference the embedded filesystem is checked against.
const migrationsDirOnDisk = "../../infrastructure/adapters/secondary/repository/migrations"

// migrationsDirAbsolute resolves the on-disk migrations directory against the
// process working directory.
//
// It must be called before anything changes that directory, and the result
// handed around as an absolute path afterwards. Writing the helper this way is
// deliberate: the first draft resolved the same directory relative to the CWD
// from inside a test that had just chdir'd into a temporary folder, and failed
// to find a single migration. That is precisely the coupling this round removes
// from the runtime, reproduced in the test that proves the removal.
func migrationsDirAbsolute(t *testing.T) string {
	t.Helper()

	dir, err := filepath.Abs(migrationsDirOnDisk)
	if err != nil {
		t.Fatalf("resolve %s against the working directory: %v", migrationsDirOnDisk, err)
	}
	return dir
}

// latestMigrationVersionOnDisk parses the highest migration number from the
// directory. Deriving the expectation from the repository rather than hardcoding
// it is the point: a new migration that is added but never embedded shows up
// here as a version mismatch instead of as a database that quietly stops short.
func latestMigrationVersionOnDisk(t *testing.T, migrationsDir string) uint {
	t.Helper()

	paths, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		t.Fatalf("list migrations under %s: %v", migrationsDir, err)
	}
	if len(paths) == 0 {
		t.Fatalf("no .up.sql migrations found under %s", migrationsDir)
	}

	var latest uint
	for _, path := range paths {
		prefix, _, ok := strings.Cut(filepath.Base(path), "_")
		if !ok {
			t.Fatalf("migration %s does not follow the <version>_<name> convention", filepath.Base(path))
		}
		number, err := strconv.ParseUint(prefix, 10, 64)
		if err != nil {
			t.Fatalf("migration %s has a non-numeric version prefix: %v", filepath.Base(path), err)
		}
		if uint(number) > latest {
			latest = uint(number)
		}
	}
	return latest
}

// scratchDatabaseURL provisions a throwaway database inside the already running
// PostgreSQL container and returns a connection string pointed at it.
//
// A separate database is what makes this safe: the runner records progress in a
// schema_migrations table, so running it against the shared fixture database
// would either collide with the init scripts or migrate the fixtures every other
// test in this package depends on.
func scratchDatabaseURL(t *testing.T, name string) string {
	t.Helper()

	ctx := context.Background()
	base, err := TestContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("container connection string: %v", err)
	}

	parsed, err := url.Parse(base)
	if err != nil {
		t.Fatalf("parse connection string %q: %v", base, err)
	}
	parsed.Path = "/" + name
	// testcontainers builds this URL with an empty query string, which is enough
	// for pgx (it falls back from TLS when the server refuses) but not for
	// lib/pq, the driver golang-migrate opens its connection with: pq defaults
	// to sslmode=require and the container has no TLS. Stating it explicitly
	// keeps the two clients in agreement.
	query := parsed.Query()
	query.Set("sslmode", "disable")
	parsed.RawQuery = query.Encode()
	scratchURL := parsed.String()

	// Drop first so a database orphaned by an aborted previous run cannot make
	// the create below fail, and drop again afterwards so nothing accumulates.
	for _, statement := range []string{
		"DROP DATABASE IF EXISTS " + name + " WITH (FORCE)",
		"CREATE DATABASE " + name,
	} {
		if _, err := TestPool.Exec(ctx, statement); err != nil {
			t.Fatalf("%q: %v", statement, err)
		}
	}
	t.Cleanup(func() {
		if _, err := TestPool.Exec(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)"); err != nil {
			t.Logf("drop scratch database %s: %v", name, err)
		}
	})

	return scratchURL
}

// TestMigrationsRunFromAnArbitraryWorkingDirectory is the proof that booting no
// longer depends on the process working directory.
//
// MakeMigrations used to read "file://infrastructure/adapters/secondary/
// repository/migrations": a path relative to the CWD, which resolved only when
// the process happened to start inside the module root. This test makes that
// condition impossible to satisfy. It changes into an empty temporary directory
// that contains none of the module's files, builds the source through
// database.NewMigrationSource — the same constructor MakeMigrations calls — and
// runs the real migration runner against a real database.
//
// A filesystem-based source would fail right here at NewMigrationSource with a
// "pattern matches no files" error, which is exactly the failure the embed
// replaces.
func TestMigrationsRunFromAnArbitraryWorkingDirectory(t *testing.T) {
	// Resolved before the chdir below, while the working directory is still the
	// one `go test` started in.
	migrationsDir := migrationsDirAbsolute(t)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("read the working directory: %v", err)
	}

	// t.TempDir registers its own cleanup first, so the restore below runs before
	// the directory is removed and every later test sees the original CWD.
	elsewhere := t.TempDir()
	if err := os.Chdir(elsewhere); err != nil {
		t.Fatalf("chdir into %s: %v", elsewhere, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore the working directory to %s: %v", originalDir, err)
		}
	})

	// Assert the precondition rather than trusting it: if the temporary directory
	// somehow had a matching tree, the test would no longer prove anything.
	if _, err := os.Stat(filepath.Join(elsewhere, "infrastructure")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the temporary directory unexpectedly holds an infrastructure tree (stat error = %v); "+
			"the test would no longer prove CWD independence", err)
	}

	src, err := database.NewMigrationSource()
	if err != nil {
		t.Fatalf("NewMigrationSource() while the working directory is %s: %v", elsewhere, err)
	}

	scratchURL := scratchDatabaseURL(t, "migration_cwd_probe")
	m, err := migrate.NewWithSourceInstance("iofs", src, scratchURL)
	if err != nil {
		t.Fatalf("build the migration runner while the working directory is %s: %v", elsewhere, err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("m.Up() while the working directory is %s: %v", elsewhere, err)
	}

	want := latestMigrationVersionOnDisk(t, migrationsDir)
	version, dirty, err := m.Version()
	if err != nil {
		t.Fatalf("read the migration version: %v", err)
	}
	if dirty {
		t.Error("the schema is dirty after a full run from a foreign working directory")
	}
	if version != want {
		t.Errorf("schema version = %d, want %d (every embedded migration must be applied)", version, want)
	}

	probe, err := pgxpool.New(context.Background(), scratchURL)
	if err != nil {
		t.Fatalf("connect to the scratch database: %v", err)
	}
	defer probe.Close()

	for _, table := range []string{
		"addresses", "supply_requests", "supply_offers", "matches", "transactions", "supplier_inventory",
	} {
		var exists bool
		if err := probe.QueryRow(context.Background(),
			`SELECT to_regclass('public.' || $1) IS NOT NULL`, table).Scan(&exists); err != nil {
			t.Fatalf("probe table %s: %v", table, err)
		}
		if !exists {
			t.Errorf("table %s is missing after migrating from a foreign working directory", table)
		}
	}

	// The newest migration is the index this round added, so finding it proves the
	// embedded set is neither truncated nor stale.
	var indexExists bool
	if err := probe.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = $1)`,
		"idx_supply_requests_status_created").Scan(&indexExists); err != nil {
		t.Fatalf("probe the newest index: %v", err)
	}
	if !indexExists {
		t.Error("idx_supply_requests_status_created is missing from the schema migrated from a foreign working directory")
	}
}

// TestEmbeddedMigrationsCoverEveryFileOnDisk ties the embedded filesystem to the
// directory it replaced.
//
// An embed pattern that silently matched a subset — a typo in the pattern, a new
// migration added without a rebuild — would sail through the rest of this
// package, because the TestContainers harness replays migrations from disk and
// never touches the embed. This is the only test that compares the two.
func TestEmbeddedMigrationsCoverEveryFileOnDisk(t *testing.T) {
	migrationsDir := migrationsDirAbsolute(t)

	onDisk, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		t.Fatalf("list the migrations directory: %v", err)
	}
	if len(onDisk) == 0 {
		t.Fatalf("no migration files found under %s", migrationsDir)
	}

	entries, err := fs.ReadDir(repository.Migrations, "migrations")
	if err != nil {
		t.Fatalf("read the embedded migrations directory: %v", err)
	}

	embedded := make(map[string]bool, len(entries))
	for _, entry := range entries {
		embedded[entry.Name()] = true
	}

	for _, path := range onDisk {
		name := filepath.Base(path)
		if !embedded[name] {
			t.Errorf("migration %s exists on disk but is missing from the embedded filesystem", name)
		}
	}
	if len(entries) != len(onDisk) {
		t.Errorf("embedded migrations = %d, on-disk migrations = %d; the embed pattern and the directory have drifted",
			len(entries), len(onDisk))
	}
}
