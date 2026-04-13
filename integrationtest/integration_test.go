//go:build integration

package integrationtest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/exc-works/sql-migrate/internal/dialect"
	"github.com/exc-works/sql-migrate/internal/logger"
	"github.com/exc-works/sql-migrate/internal/migrate"
	"github.com/exc-works/sql-migrate/internal/source"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMigrateIntegration(t *testing.T) {
	requireProviderHealthy(t)
	for _, flavor := range selectedFlavors() {
		flavor := flavor
		t.Run(flavor, func(t *testing.T) {
			t.Parallel()
			db, d, cleanup := startContainerDB(t, flavor)
			defer cleanup()
			runScenario(t, db, d)
		})
	}
}

func requireProviderHealthy(t *testing.T) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if isCI() {
				t.Fatalf("testcontainers provider panic in CI: %v", r)
			}
			t.Skipf("testcontainers provider unavailable: %v", r)
		}
	}()

	ctx := context.Background()
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err == nil {
		err = provider.Health(ctx)
	}
	if err != nil {
		if isCI() {
			t.Fatalf("testcontainers provider unhealthy in CI: %v", err)
		}
		t.Skipf("testcontainers provider unavailable: %v", err)
	}
}

func isCI() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("CI")), "true") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("GITHUB_ACTIONS")), "true")
}

func runScenario(t *testing.T, db *sql.DB, d dialect.Dialect) {
	t.Helper()
	ctx := context.Background()
	migrationDir := filepath.Join("testdata")

	primarySchema := randomName("migration_schema")
	svc, err := migrate.NewService(ctx, migrate.Config{
		Dialect:           d,
		DB:                db,
		Logger:            logger.NoopLogger{},
		SchemaName:        primarySchema,
		MigrateOutOfOrder: false,
		MigrationSource:   source.DirectorySource{Directory: migrationDir},
	})
	if err != nil {
		t.Fatalf("new service: %v", err)
	}

	if err := svc.Create(); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	statuses, err := svc.Status()
	if err != nil {
		t.Fatalf("status before up: %v", err)
	}
	if len(statuses) != 3 {
		t.Fatalf("unexpected status count: %d", len(statuses))
	}

	if err := svc.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}
	statuses, err = svc.Status()
	if err != nil {
		t.Fatalf("status after up: %v", err)
	}
	for _, st := range statuses {
		if st.Status != migrate.StatusApplied {
			t.Fatalf("expected applied, got %s for %s", st.Status, st.Migration.Filename)
		}
	}

	// up idempotence
	if err := svc.Up(); err != nil {
		t.Fatalf("second up: %v", err)
	}

	// down to version 2 should rollback only V3
	if err := svc.Down("2", false); err != nil {
		t.Fatalf("down to version 2: %v", err)
	}
	count := queryCount(t, db, "SELECT COUNT(*) FROM accounts")
	if count != 1 {
		t.Fatalf("expected 1 account after down to version 2, got %d", count)
	}

	// baseline scenario
	baselineSchema := randomName("migration_schema")
	baselineSvc, err := migrate.NewService(ctx, migrate.Config{
		Dialect:           d,
		DB:                db,
		Logger:            logger.NoopLogger{},
		SchemaName:        baselineSchema,
		MigrateOutOfOrder: false,
		MigrationSource:   source.DirectorySource{Directory: migrationDir},
	})
	if err != nil {
		t.Fatalf("new baseline service: %v", err)
	}
	if err := baselineSvc.Create(); err != nil {
		t.Fatalf("baseline create: %v", err)
	}
	if err := baselineSvc.Baseline(); err != nil {
		t.Fatalf("baseline: %v", err)
	}
	baselineStatuses, err := baselineSvc.Status()
	if err != nil {
		t.Fatalf("baseline status: %v", err)
	}
	for _, st := range baselineStatuses {
		if st.Status != migrate.StatusBaseline {
			t.Fatalf("expected baseline status, got %s for %s", st.Status, st.Migration.Filename)
		}
	}

	// hash mismatch scenario (modify V2 content)
	tamperedDir := t.TempDir()
	copyMigrationDir(t, migrationDir, tamperedDir)
	file := filepath.Join(tamperedDir, "V2__seed_accounts.sql")
	content, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read tamper file: %v", err)
	}
	content = append(content, []byte("\n-- tampered\n")...)
	if err := os.WriteFile(file, content, 0o644); err != nil {
		t.Fatalf("write tamper file: %v", err)
	}
	mismatchSvc, err := migrate.NewService(ctx, migrate.Config{
		Dialect:         d,
		DB:              db,
		Logger:          logger.NoopLogger{},
		SchemaName:      baselineSchema,
		MigrationSource: source.DirectorySource{Directory: tamperedDir},
	})
	if err != nil {
		t.Fatalf("new mismatch service: %v", err)
	}
	if err := mismatchSvc.Up(); err == nil {
		t.Fatalf("expected hash mismatch error")
	}
}

func queryCount(t *testing.T, db *sql.DB, query string) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		t.Fatalf("query count failed: %v", err)
	}
	return count
}

func selectedFlavors() []string {
	v := strings.TrimSpace(os.Getenv("INTEGRATION_DB"))
	if v == "" {
		return []string{"postgres", "mysql", "mariadb"}
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"postgres", "mysql", "mariadb"}
	}
	return out
}

func startContainerDB(t *testing.T, flavor string) (*sql.DB, dialect.Dialect, func()) {
	t.Helper()
	ctx := context.Background()

	var (
		request    testcontainers.ContainerRequest
		port       string
		driverName string
		dsn        func(host, mappedPort string) string
		d          dialect.Dialect
	)
	dbName := "testdb"
	userName := "tc_user"
	userPass := randomSecret()
	rootPass := randomSecret()

	switch flavor {
	case "postgres":
		request = testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       dbName,
				"POSTGRES_USER":     userName,
				"POSTGRES_PASSWORD": userPass,
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("5432/tcp"),
				wait.ForLog("database system is ready to accept connections"),
			),
		}
		port = "5432/tcp"
		driverName = "pgx"
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, mappedPort, userName, userPass, dbName)
		}
		d = dialect.PostgresDialect{}
	case "mysql":
		request = testcontainers.ContainerRequest{
			Image:        "mysql:8.0",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MYSQL_DATABASE":      dbName,
				"MYSQL_USER":          userName,
				"MYSQL_PASSWORD":      userPass,
				"MYSQL_ROOT_PASSWORD": rootPass,
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("3306/tcp"),
				wait.ForLog("ready for connections"),
			),
		}
		port = "3306/tcp"
		driverName = "mysql"
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", userName, userPass, host, mappedPort, dbName)
		}
		d = dialect.MySQLDialect{}
	case "mariadb":
		request = testcontainers.ContainerRequest{
			Image:        "mariadb:11",
			ExposedPorts: []string{"3306/tcp"},
			Env: map[string]string{
				"MARIADB_DATABASE":      dbName,
				"MARIADB_USER":          userName,
				"MARIADB_PASSWORD":      userPass,
				"MARIADB_ROOT_PASSWORD": rootPass,
			},
			WaitingFor: wait.ForAll(
				wait.ForListeningPort("3306/tcp"),
				wait.ForLog("ready for connections"),
			),
		}
		port = "3306/tcp"
		driverName = "mysql"
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", userName, userPass, host, mappedPort, dbName)
		}
		d = dialect.MariaDBDialect{}
	default:
		t.Fatalf("unsupported flavor: %s", flavor)
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start %s container: %v", flavor, err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("resolve host: %v", err)
	}
	mapped, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("resolve port: %v", err)
	}
	db, err := sql.Open(driverName, dsn(host, mapped.Port()))
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("open db: %v", err)
	}

	deadline := time.Now().Add(60 * time.Second)
	for {
		err = db.PingContext(ctx)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = db.Close()
			_ = container.Terminate(ctx)
			t.Fatalf("ping db timeout: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}
	return db, d, cleanup
}

func randomName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}

func randomSecret() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("tc_%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func copyMigrationDir(t *testing.T, from, to string) {
	t.Helper()
	entries, err := os.ReadDir(from)
	if err != nil {
		t.Fatalf("read source dir: %v", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(from, entry.Name())
		dst := filepath.Join(to, entry.Name())
		content, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if err := os.WriteFile(dst, content, 0o644); err != nil {
			t.Fatalf("write %s: %v", dst, err)
		}
	}
}
