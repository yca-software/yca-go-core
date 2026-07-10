package yca_postgresql_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	yca_postgresql "github.com/yca-software/yca-go-core/postgresql"
)

type PostgreSQLSuite struct {
	suite.Suite
	postgresContainer testcontainers.Container
	testDSN           string
}

func TestPostgreSQLSuite(t *testing.T) {
	suite.Run(t, new(PostgreSQLSuite))
}

func (s *PostgreSQLSuite) SetupSuite() {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	s.Require().NoError(err)
	s.postgresContainer = container

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)
	s.testDSN = connStr
}

func (s *PostgreSQLSuite) TearDownSuite() {
	if s.postgresContainer == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.NoError(s.postgresContainer.Terminate(ctx))
}

func (s *PostgreSQLSuite) TestNewPostgreSQL_invalidDSN() {
	pg, err := yca_postgresql.NewPostgreSQL(yca_postgresql.PostgreSQLClientConfig{
		DSN: "postgres://invalid:invalid@127.0.0.1:1/nope?sslmode=disable",
	})
	s.Error(err)
	s.Nil(pg)
}

func (s *PostgreSQLSuite) TestNewPostgreSQL_success() {
	pg, err := yca_postgresql.NewPostgreSQL(yca_postgresql.PostgreSQLClientConfig{
		DSN:             s.testDSN,
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		ConnMaxIdleTime: 30 * time.Second,
	})
	s.Require().NoError(err)
	s.Require().NotNil(pg)
	s.Require().NotNil(pg.GetClient())

	ctx := context.Background()
	s.NoError(pg.Check(ctx))

	pg.Cleanup()
	s.Error(pg.Check(ctx))
}

func (s *PostgreSQLSuite) TestNewPostgreSQL_Check() {
	pg, err := yca_postgresql.NewPostgreSQL(yca_postgresql.PostgreSQLClientConfig{
		DSN:          s.testDSN,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})
	s.Require().NoError(err)
	defer pg.Cleanup()

	ctx := context.Background()
	s.NoError(pg.Check(ctx))
}

func (s *PostgreSQLSuite) TestNewPostgreSQL_Check_cancelledContext() {
	pg, err := yca_postgresql.NewPostgreSQL(yca_postgresql.PostgreSQLClientConfig{
		DSN:          s.testDSN,
		MaxOpenConns: 5,
		MaxIdleConns: 2,
	})
	s.Require().NoError(err)
	defer pg.Cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	s.Error(pg.Check(ctx))
}
