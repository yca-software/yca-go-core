package yca_test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	yca_database "github.com/yca-software/yca-go-core/postgresql"
)

// testWaitTimeout must cover first-boot on slow Docker Desktop hosts.
const testWaitTimeout = 5 * time.Minute

// ContainerStartTimeout bounds image pull + container start for testcontainers.
const ContainerStartTimeout = 5 * time.Minute

// PostgresImage ships PostGIS (migrations/00001_extensions.sql). Pin the tag for reproducible CI.
const PostgresImage = "postgis/postgis:16-3.4"

// EnvIntegrationTestPostgresImage overrides the testcontainers Postgres image
// (e.g. 2chi-postgres:16-postgis-pgvector when migrations need pgvector).
const EnvIntegrationTestPostgresImage = "GO_API_INTEGRATION_TEST_POSTGRES_IMAGE"

// integrationContainerName is used with testcontainers reuse so repository test processes share one DB.
// Run repository integration tests with -p 1 (see Makefile).
const integrationContainerName = "2chi-go-api-integration-postgis"

// EnvIntegrationTestDSN skips testcontainers when set
// (e.g. postgres://user:pass@localhost:5432/testdb?sslmode=disable).
const EnvIntegrationTestDSN = "GO_API_INTEGRATION_TEST_DSN"

// EnvIntegrationTestNoContainerReuse forces a fresh container per process when set to "1".
const EnvIntegrationTestNoContainerReuse = "GO_API_INTEGRATION_TEST_NO_CONTAINER_REUSE"

var (
	testOnce          sync.Once
	testInstance      *DB
	testSetupErr      error
	testMu            sync.Mutex
	testMigrationsDir string
)

// DB is a migrated Postgres used in repository integration tests.
type DB struct {
	container          testcontainers.Container
	sqlDB              *sql.DB
	connStr            string
	terminateContainer bool
}

// Get returns a shared test database (container or EnvIntegrationTestDSN).
// migrationsDir is the goose migrations directory for the calling module (absolute or relative to the test process cwd).
func Get(migrationsDir string) (*DB, error) {
	if strings.TrimSpace(migrationsDir) == "" {
		return nil, fmt.Errorf("testdb: migrations dir is required")
	}
	testMigrationsDir = migrationsDir
	testOnce.Do(func() {
		testInstance, testSetupErr = setup()
	})
	return testInstance, testSetupErr
}

func postgresImage() string {
	if img := strings.TrimSpace(os.Getenv(EnvIntegrationTestPostgresImage)); img != "" {
		return img
	}
	return PostgresImage
}

func setup() (*DB, error) {
	if dsn := strings.TrimSpace(os.Getenv(EnvIntegrationTestDSN)); dsn != "" {
		ctx, cancel := context.WithTimeout(context.Background(), testWaitTimeout)
		defer cancel()
		return openAndMigrate(ctx, dsn, nil, false)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ContainerStartTimeout)
	defer cancel()

	opts := []testcontainers.ContainerCustomizer{
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForAll(
				wait.ForListeningPort("5432/tcp").SkipInternalCheck(),
				wait.ForLog("database system is ready to accept connections").WithOccurrence(1),
			).WithStartupTimeout(ContainerStartTimeout),
		),
	}

	terminateContainer := true
	if os.Getenv(EnvIntegrationTestNoContainerReuse) != "1" {
		opts = append(opts, testcontainers.WithReuseByName(integrationContainerName))
		terminateContainer = false
	}

	container, err := postgres.Run(ctx, postgresImage(), opts...)
	if err != nil {
		return nil, fmt.Errorf("testdb: start container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		terminate(ctx, container)
		return nil, fmt.Errorf("testdb: connection string: %w", err)
	}

	return openAndMigrate(ctx, connStr, container, terminateContainer)
}

func openAndMigrate(ctx context.Context, connStr string, container testcontainers.Container, terminateContainer bool) (*DB, error) {
	pg, err := waitForPostgres(ctx, connStr)
	if err != nil {
		if container != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			terminate(ctx, container)
		}
		return nil, fmt.Errorf("testdb: connect: %w", err)
	}

	sqlDB := pg.GetClient().(*sqlx.DB).DB
	if err := goose.Up(sqlDB, testMigrationsDir); err != nil {
		pg.Cleanup()
		if container != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			terminate(ctx, container)
		}
		return nil, fmt.Errorf("testdb: migrate: %w", err)
	}

	return &DB{
		container:          container,
		sqlDB:              sqlDB,
		connStr:            connStr,
		terminateContainer: terminateContainer,
	}, nil
}

// SQLDB returns the underlying *sql.DB (migrations applied).
func (t *DB) SQLDB() *sql.DB {
	return t.sqlDB
}

// SQLx returns a new sqlx handle on the same DSN (preferred for repositories).
func (t *DB) SQLx() (*sqlx.DB, error) {
	return sqlx.Connect("pgx", t.connStr)
}

// ConnectionString returns the Postgres DSN.
func (t *DB) ConnectionString() string {
	return t.connStr
}

// Cleanup closes connections and terminates the container when this process started it.
func Cleanup() {
	testMu.Lock()
	defer testMu.Unlock()
	if testInstance == nil {
		return
	}
	if testInstance.sqlDB != nil {
		_ = testInstance.sqlDB.Close()
	}
	if testInstance.container != nil && testInstance.terminateContainer {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = testInstance.container.Terminate(ctx)
	}
	testInstance = nil
	testSetupErr = nil
	testMigrationsDir = ""
	testOnce = sync.Once{}
}

func waitForPostgres(ctx context.Context, connStr string) (*yca_database.PostgreSQL, error) {
	deadline, ok := ctx.Deadline()
	if !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, testWaitTimeout)
		defer cancel()
		deadline, _ = ctx.Deadline()
	}

	var lastErr error
	for time.Now().Before(deadline) {
		pg, err := yca_database.NewPostgreSQL(yca_database.PostgreSQLClientConfig{
			DSN:          connStr,
			MaxOpenConns: 2,
			MaxIdleConns: 1,
		})
		if err != nil {
			lastErr = err
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := pg.Check(ctx); err != nil {
			lastErr = err
			pg.Cleanup()
			time.Sleep(500 * time.Millisecond)
			continue
		}
		return pg, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("testdb: postgres not ready: %w", lastErr)
	}
	return nil, fmt.Errorf("testdb: postgres not ready before deadline")
}

func terminate(ctx context.Context, container testcontainers.Container) {
	_ = container.Terminate(ctx)
}
