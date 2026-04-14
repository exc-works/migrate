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
	"sync"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/exc-works/sql-migrate/internal/dialect"
	"github.com/exc-works/sql-migrate/internal/logger"
	"github.com/exc-works/sql-migrate/internal/migrate"
	"github.com/exc-works/sql-migrate/internal/source"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/sijms/go-ora/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "modernc.org/sqlite"
)

func TestMigrateIntegration(t *testing.T) {
	for _, flavor := range selectedFlavors() {
		flavor := flavor
		t.Run(flavor, func(t *testing.T) {
			t.Parallel()
			db, d, cleanup := startDB(t, flavor)
			defer cleanup()
			runScenario(t, db, d, migrationDirForFlavor(flavor))
		})
	}
}

var providerHealth struct {
	once     sync.Once
	err      error
	panicVal any
}

func requireProviderHealthy(t *testing.T) {
	t.Helper()

	providerHealth.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				providerHealth.panicVal = r
			}
		}()
		ctx := context.Background()
		provider, err := testcontainers.ProviderDocker.GetProvider()
		if err == nil {
			err = provider.Health(ctx)
		}
		providerHealth.err = err
	})

	if providerHealth.panicVal != nil {
		if isCI() {
			t.Fatalf("testcontainers provider panic in CI: %v", providerHealth.panicVal)
		}
		t.Skipf("testcontainers provider unavailable: %v", providerHealth.panicVal)
	}

	if providerHealth.err != nil {
		if isCI() {
			t.Fatalf("testcontainers provider unhealthy in CI: %v", providerHealth.err)
		}
		t.Skipf("testcontainers provider unavailable: %v", providerHealth.err)
	}
}

func isCI() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("CI")), "true") ||
		strings.EqualFold(strings.TrimSpace(os.Getenv("GITHUB_ACTIONS")), "true")
}

func runScenario(t *testing.T, db *sql.DB, d dialect.Dialect, migrationDir string) {
	t.Helper()
	ctx := context.Background()

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

const (
	envIntegrationDB      = "INTEGRATION_DB"
	envIntegrationSQLite  = "INTEGRATION_SQLITE_DSN"
	envIntegrationOracle  = "INTEGRATION_ORACLE_DSN"
	defaultPingTimeoutSec = 60
)

var defaultIntegrationFlavors = []string{"postgres", "mysql", "mariadb", "sqlite", "oracle"}

func selectedFlavors() []string {
	v := strings.TrimSpace(os.Getenv(envIntegrationDB))
	if v == "" {
		return append([]string(nil), defaultIntegrationFlavors...)
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			if p == "all" {
				return append([]string(nil), defaultIntegrationFlavors...)
			}
			if _, ok := seen[p]; ok {
				continue
			}
			seen[p] = struct{}{}
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return append([]string(nil), defaultIntegrationFlavors...)
	}
	return out
}

func migrationDirForFlavor(flavor string) string {
	switch flavor {
	case "oracle":
		return filepath.Join("testdata_oracle")
	case "sqlite":
		return filepath.Join("testdata_sqlite")
	default:
		return filepath.Join("testdata")
	}
}

func startDB(t *testing.T, flavor string) (*sql.DB, dialect.Dialect, func()) {
	t.Helper()

	d, err := dialect.FromName(flavor)
	if err != nil {
		if isOptionalFlavor(flavor) {
			t.Skipf("%s integration skipped: dialect not available: %v", flavor, err)
		}
		t.Fatalf("resolve dialect %s: %v", flavor, err)
	}

	switch flavor {
	case "postgres", "mysql", "mariadb":
		db, cleanup := startContainerDB(t, flavor, d)
		return db, d, cleanup
	case "sqlite":
		db, cleanup := startSQLiteDB(t, d)
		return db, d, cleanup
	case "oracle":
		db, cleanup := startOracleDB(t, d)
		return db, d, cleanup
	default:
		t.Fatalf("unsupported flavor: %s", flavor)
	}
	return nil, nil, func() {}
}

func startContainerDB(t *testing.T, flavor string, d dialect.Dialect) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()
	requireProviderHealthy(t)

	var (
		request testcontainers.ContainerRequest
		port    string
		dsn     func(host, mappedPort string) string
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
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, mappedPort, userName, userPass, dbName)
		}
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
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", userName, userPass, host, mappedPort, dbName)
		}
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
		dsn = func(host, mappedPort string) string {
			return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true", userName, userPass, host, mappedPort, dbName)
		}
	default:
		t.Fatalf("unsupported flavor: %s", flavor)
	}

	driverName := requireDriver(t, flavor, false, d.DriverName())

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
	waitForPing(t, db, defaultPingTimeoutSec*time.Second)

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}
	return db, cleanup
}

func startSQLiteDB(t *testing.T, d dialect.Dialect) (*sql.DB, func()) {
	t.Helper()

	driverName := requireDriver(t, "sqlite", true, d.DriverName(), "sqlite3")
	dsn := strings.TrimSpace(os.Getenv(envIntegrationSQLite))
	if dsn == "" {
		dsn = filepath.Join(t.TempDir(), "integration.sqlite")
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	waitForPing(t, db, 10*time.Second)
	return db, func() {
		_ = db.Close()
	}
}

func startOracleDB(t *testing.T, d dialect.Dialect) (*sql.DB, func()) {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv(envIntegrationOracle))
	if dsn == "" {
		t.Skipf("oracle integration skipped: %s is not set", envIntegrationOracle)
	}

	driverName := requireDriver(t, "oracle", true, d.DriverName(), "godror")
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		t.Fatalf("open oracle db: %v", err)
	}
	waitForPing(t, db, defaultPingTimeoutSec*time.Second)
	cleanupOracleAccountsTable(t, db)
	return db, func() {
		_ = db.Close()
	}
}

func cleanupOracleAccountsTable(t *testing.T, db *sql.DB) {
	t.Helper()
	const stmt = `
BEGIN
    EXECUTE IMMEDIATE 'DROP TABLE accounts';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -942 THEN
            RAISE;
        END IF;
END;
`
	if _, err := db.ExecContext(context.Background(), stmt); err != nil {
		t.Fatalf("oracle pre-clean failed: %v", err)
	}
}

func waitForPing(t *testing.T, db *sql.DB, timeout time.Duration) {
	t.Helper()
	ctx := context.Background()
	deadline := time.Now().Add(timeout)
	var err error
	for {
		err = db.PingContext(ctx)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("ping db timeout: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func requireDriver(t *testing.T, flavor string, allowSkip bool, names ...string) string {
	t.Helper()
	driver := firstRegisteredDriver(names...)
	if driver != "" {
		return driver
	}

	checked := strings.Join(uniqueNonEmpty(names), ", ")
	if allowSkip {
		t.Skipf("%s integration skipped: SQL driver not registered (checked: %s)", flavor, checked)
	}
	t.Fatalf("%s integration setup failed: SQL driver not registered (checked: %s)", flavor, checked)
	return ""
}

func firstRegisteredDriver(names ...string) string {
	drivers := sql.Drivers()
	for _, candidate := range names {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		for _, registered := range drivers {
			if registered == candidate {
				return candidate
			}
		}
	}
	return ""
}

func uniqueNonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func isOptionalFlavor(flavor string) bool {
	switch flavor {
	case "sqlite", "oracle":
		return true
	default:
		return false
	}
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
