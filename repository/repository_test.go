package chi_repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	chi_error "github.com/yca-software/yca-go-core/error"
	chi_observer "github.com/yca-software/yca-go-core/observer"
	chi_repository "github.com/yca-software/yca-go-core/repository"
)

type Product struct {
	ID          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Description *string   `db:"description"`
	Price       float64   `db:"price"`
	Stock       int       `db:"stock"`
	Category    *string   `db:"category"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type mockMetricsHook struct {
	records []metricRecord
}

type metricRecord struct {
	operation string
	table     string
	queryType string
	status    string
	duration  time.Duration
}

func (m *mockMetricsHook) RecordQuery(operation, table, queryType, status string, duration time.Duration) {
	m.records = append(m.records, metricRecord{
		operation: operation,
		table:     table,
		queryType: queryType,
		status:    status,
		duration:  duration,
	})
}

func (m *mockMetricsHook) reset() {
	m.records = nil
}

type RepositorySuite struct {
	suite.Suite
	db              *sqlx.DB
	container       testcontainers.Container
	repo            chi_repository.Repository[Product]
	repoWithMetrics chi_repository.Repository[Product]
	metricsHook     *mockMetricsHook
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}

func (s *RepositorySuite) SetupSuite() {
	ctx := context.Background()

	container, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	s.Require().NoError(err)
	s.container = container

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	s.Require().NoError(err)

	config, err := pgx.ParseConfig(connStr)
	s.Require().NoError(err)

	s.db = sqlx.NewDb(stdlib.OpenDB(*config), "pgx")
	s.runMigrations()

	columns := []string{"id", "name", "description", "price", "stock", "category", "is_active", "created_at", "updated_at"}
	s.repo = chi_repository.NewRepository[Product](s.db, "products", columns, nil)

	s.metricsHook = &mockMetricsHook{}
	var hook chi_observer.QueryMetricsHook = s.metricsHook
	s.repoWithMetrics = chi_repository.NewRepository[Product](s.db, "products", columns, hook)
}

func (s *RepositorySuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s.NoError(s.container.Terminate(ctx))
}

func (s *RepositorySuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE products RESTART IDENTITY CASCADE")
	s.Require().NoError(err)
	s.metricsHook.reset()
}

func (s *RepositorySuite) runMigrations() {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			stock INTEGER NOT NULL DEFAULT 0,
			category VARCHAR(100),
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name)
		);
	`)
	s.Require().NoError(err)
}

func (s *RepositorySuite) TestCreateAndGet() {
	ctx := context.Background()
	err := s.repo.Create(ctx, map[string]any{
		"name": "Test Product", "price": 99.99, "stock": 10, "is_active": true,
	})
	s.Require().NoError(err)

	product, err := s.repo.Get(ctx, squirrel.Eq{"name": "Test Product"}, nil)
	s.Require().NoError(err)
	s.Equal("Test Product", product.Name)
	s.Equal(99.99, product.Price)
}

func (s *RepositorySuite) TestGet_NilCondition_ReturnsError() {
	_, err := s.repo.Get(context.Background(), nil, nil)
	s.ErrorIs(err, chi_repository.ErrConditionRequired)
}

func (s *RepositorySuite) TestGet_NotFound() {
	product, err := s.repo.Get(context.Background(), squirrel.Eq{"name": "missing"}, nil)
	s.Error(err)
	s.Nil(product)

	var apiErr *chi_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(404, apiErr.StatusCode)
}

func (s *RepositorySuite) TestSelect_WithSort() {
	ctx := context.Background()
	for _, p := range []map[string]any{
		{"name": "Product 1", "price": 10.0, "stock": 1, "is_active": true, "category": "cat1"},
		{"name": "Product 2", "price": 20.0, "stock": 2, "is_active": true, "category": "cat1"},
	} {
		s.Require().NoError(s.repo.Create(ctx, p))
	}

	results, err := s.repo.Select(ctx, squirrel.Eq{"category": "cat1"}, nil, "price ASC")
	s.Require().NoError(err)
	s.Require().Len(*results, 2)
	s.Equal("Product 1", (*results)[0].Name)
}

func (s *RepositorySuite) TestSelect_InvalidSortRejected() {
	_, err := s.repo.Select(context.Background(), nil, nil, "price; DROP TABLE products; --")
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid sort")
}

func (s *RepositorySuite) TestPaginatedSelect() {
	ctx := context.Background()
	for i := 1; i <= 5; i++ {
		s.Require().NoError(s.repo.Create(ctx, map[string]any{
			"name": fmt.Sprintf("Product %d", i), "price": float64(i * 10), "stock": i, "is_active": true,
		}))
	}

	results, err := s.repo.PaginatedSelect(ctx, nil, nil, "price ASC", 2, 0)
	s.Require().NoError(err)
	s.Len(*results, 2)
	s.Equal("Product 1", (*results)[0].Name)
}

func (s *RepositorySuite) TestPaginatedSelect_RequiresSort() {
	_, err := s.repo.PaginatedSelect(context.Background(), nil, nil, "", 10, 0)
	s.Require().Error(err)
	s.Contains(err.Error(), "sort parameter is required")
}

func (s *RepositorySuite) TestCount() {
	ctx := context.Background()
	for _, p := range []map[string]any{
		{"name": "Count 1", "price": 10.0, "stock": 1, "is_active": true},
		{"name": "Count 2", "price": 20.0, "stock": 2, "is_active": true},
		{"name": "Count 3", "price": 30.0, "stock": 3, "is_active": false},
	} {
		s.Require().NoError(s.repo.Create(ctx, p))
	}

	count, err := s.repo.Count(ctx, nil)
	s.Require().NoError(err)
	s.Equal(3, count)

	count, err = s.repo.Count(ctx, squirrel.Eq{"is_active": true})
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *RepositorySuite) TestUpdate() {
	ctx := context.Background()
	s.Require().NoError(s.repo.Create(ctx, map[string]any{
		"name": "Update Product", "price": 50.0, "stock": 5, "is_active": true,
	}))

	s.Require().NoError(s.repo.Update(ctx, squirrel.Eq{"name": "Update Product"}, map[string]any{
		"price": 75.0,
	}))

	product, err := s.repo.Get(ctx, squirrel.Eq{"name": "Update Product"}, nil)
	s.Require().NoError(err)
	s.Equal(75.0, product.Price)
}

func (s *RepositorySuite) TestUpdate_NilCondition_ReturnsError() {
	err := s.repo.Update(context.Background(), nil, map[string]any{"price": 10.0})
	s.ErrorIs(err, chi_repository.ErrConditionRequired)
}

func (s *RepositorySuite) TestDelete() {
	ctx := context.Background()
	s.Require().NoError(s.repo.Create(ctx, map[string]any{
		"name": "Delete Product", "price": 25.0, "stock": 3, "is_active": true,
	}))

	s.Require().NoError(s.repo.Delete(ctx, squirrel.Eq{"name": "Delete Product"}))

	product, err := s.repo.Get(ctx, squirrel.Eq{"name": "Delete Product"}, nil)
	s.Error(err)
	s.Nil(product)
}

func (s *RepositorySuite) TestDelete_NotFound() {
	err := s.repo.Delete(context.Background(), squirrel.Eq{"name": "missing"})
	s.Require().Error(err)

	var apiErr *chi_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(404, apiErr.StatusCode)
}

func (s *RepositorySuite) TestCreateMany() {
	ctx := context.Background()
	columns := []string{"name", "price", "stock", "is_active"}
	data := []map[string]any{
		{"name": "Bulk 1", "price": 10.0, "stock": 1, "is_active": true},
		{"name": "Bulk 2", "price": 20.0, "stock": 2, "is_active": true},
	}
	s.Require().NoError(s.repo.CreateMany(ctx, columns, data, false))

	count, err := s.repo.Count(ctx, nil)
	s.Require().NoError(err)
	s.Equal(2, count)
}

func (s *RepositorySuite) TestCreateMany_EmptyNoOp() {
	ctx := context.Background()
	s.NoError(s.repo.CreateMany(ctx, nil, nil, false))
	s.NoError(s.repo.CreateMany(ctx, []string{"name"}, nil, false))

	count, err := s.repo.Count(ctx, nil)
	s.Require().NoError(err)
	s.Equal(0, count)
}

func (s *RepositorySuite) TestUniqueConstraintError() {
	ctx := context.Background()
	data := map[string]any{"name": "Unique Product", "price": 50.0, "stock": 5, "is_active": true}
	s.Require().NoError(s.repo.Create(ctx, data))

	err := s.repo.Create(ctx, data)
	s.Require().Error(err)

	var apiErr *chi_error.Error
	s.Require().ErrorAs(err, &apiErr)
	s.Equal(409, apiErr.StatusCode)
}

func (s *RepositorySuite) TestRunInTx_Commit() {
	ctx := context.Background()
	err := chi_repository.RunInTx(ctx, s.db, nil, func(tx chi_repository.Tx) error {
		txRepo := s.repo.WithTx(tx)
		return txRepo.Create(ctx, map[string]any{
			"name": "Tx Product", "price": 100.0, "stock": 10, "is_active": true,
		})
	})
	s.Require().NoError(err)

	product, err := s.repo.Get(ctx, squirrel.Eq{"name": "Tx Product"}, nil)
	s.Require().NoError(err)
	s.Equal("Tx Product", product.Name)
}

func (s *RepositorySuite) TestRunInTx_Rollback() {
	ctx := context.Background()
	err := chi_repository.RunInTx(ctx, s.db, nil, func(tx chi_repository.Tx) error {
		txRepo := s.repo.WithTx(tx)
		if err := txRepo.Create(ctx, map[string]any{
			"name": "Rollback Product", "price": 100.0, "stock": 10, "is_active": true,
		}); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	s.Require().Error(err)

	product, err := s.repo.Get(ctx, squirrel.Eq{"name": "Rollback Product"}, nil)
	s.Error(err)
	s.Nil(product)
}

func (s *RepositorySuite) TestMetricsHook() {
	ctx := context.Background()
	s.Require().NoError(s.repoWithMetrics.Create(ctx, map[string]any{
		"name": "Metrics Product", "price": 50.0, "stock": 5, "is_active": true,
	}))

	found := false
	for _, record := range s.metricsHook.records {
		if record.operation == "exec" && record.table == "products" && record.status == "success" {
			found = true
			s.Greater(record.duration, time.Duration(0))
		}
	}
	s.True(found)
}

func (s *RepositorySuite) TestGetQueryBuilder_UsesDollarPlaceholders() {
	sqlStr, args, err := s.repo.GetQueryBuilder().Select("id").From("products").Where(squirrel.Eq{"id": 1}).ToSql()
	s.Require().NoError(err)
	s.Contains(sqlStr, "$1")
	s.Len(args, 1)
}

func (s *RepositorySuite) TestDB() {
	s.NotNil(s.repo.DB())
}
