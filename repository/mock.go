package chi_repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/mock"
)

type NoopResult struct{}

func (NoopResult) LastInsertId() (int64, error) { return 0, nil }
func (NoopResult) RowsAffected() (int64, error) { return 0, nil }

// --- Mock Tx ---

type MockTx struct {
	mock.Mock
}

func NewMockTx() *MockTx { return &MockTx{} }

func (m *MockTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return NoopResult{}, nil
}
func (m *MockTx) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *MockTx) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *MockTx) Commit() error   { return m.Called().Error(0) }
func (m *MockTx) Rollback() error { return m.Called().Error(0) }

// --- Mock Repository ---

type MockRepository[T any] struct {
	mock.Mock
}

func NewMockRepository[T any]() *MockRepository[T] {
	return &MockRepository[T]{}
}

func (m *MockRepository[T]) DB() Executor {
	return m.Called().Get(0).(Executor)
}

func (m *MockRepository[T]) WithTx(tx Tx) Repository[T] {
	return m.Called(tx).Get(0).(Repository[T])
}

func (m *MockRepository[T]) GetQueryBuilder() squirrel.StatementBuilderType {
	return m.Called().Get(0).(squirrel.StatementBuilderType)
}

func (m *MockRepository[T]) Count(ctx context.Context, condition squirrel.Sqlizer) (int, error) {
	args := m.Called(ctx, condition)
	return args.Int(0), args.Error(1)
}

func (m *MockRepository[T]) Get(ctx context.Context, condition squirrel.Sqlizer, columns *[]string) (*T, error) {
	args := m.Called(ctx, condition, columns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*T), args.Error(1)
}

func (m *MockRepository[T]) Select(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error) {
	args := m.Called(ctx, condition, columns, sort)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]T), args.Error(1)
}

func (m *MockRepository[T]) PaginatedSelect(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error) {
	args := m.Called(ctx, condition, columns, sort, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]T), args.Error(1)
}

func (m *MockRepository[T]) Create(ctx context.Context, data map[string]any) error {
	return m.Called(ctx, data).Error(0)
}

func (m *MockRepository[T]) CreateMany(ctx context.Context, columns []string, data []map[string]any, ignoreConflict bool) error {
	return m.Called(ctx, columns, data, ignoreConflict).Error(0)
}

func (m *MockRepository[T]) Delete(ctx context.Context, condition squirrel.Sqlizer) error {
	return m.Called(ctx, condition).Error(0)
}

func (m *MockRepository[T]) Update(ctx context.Context, condition squirrel.Sqlizer, data map[string]any) error {
	return m.Called(ctx, condition, data).Error(0)
}

var _ Repository[any] = (*MockRepository[any])(nil)
