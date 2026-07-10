package yca_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	yca_observer "github.com/yca-software/yca-go-core/observer"
)

var ErrConditionRequired = errors.New("repository: condition is required for Get, Delete, and Update")

// Executor mandates context for all database interactions.
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

// Tx represents an active database transaction.
type Tx interface {
	Executor
	Commit() error
	Rollback() error
}

type Repository[T any] interface {
	DB() Executor
	WithTx(tx Tx) Repository[T]
	GetQueryBuilder() squirrel.StatementBuilderType

	Count(ctx context.Context, condition squirrel.Sqlizer) (int, error)
	Get(ctx context.Context, condition squirrel.Sqlizer, columns *[]string) (*T, error)
	Select(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error)
	PaginatedSelect(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error)
	Create(ctx context.Context, data map[string]any) error
	CreateMany(ctx context.Context, columns []string, data []map[string]any, ignoreConflict bool) error
	Delete(ctx context.Context, condition squirrel.Sqlizer) error
	Update(ctx context.Context, condition squirrel.Sqlizer, data map[string]any) error
}

type repository[T any] struct {
	exec      Executor
	tableName string
	columns   []string
	metrics   yca_observer.QueryMetricsHook
}

func NewRepository[T any](db *sqlx.DB, tableName string, columns []string, hook yca_observer.QueryMetricsHook) Repository[T] {
	if db == nil {
		panic("repository: db must not be nil")
	}
	if tableName == "" {
		panic("repository: table name must not be empty")
	}
	if len(columns) == 0 {
		panic("repository: columns must not be empty")
	}

	return &repository[T]{
		exec:      NewDBWrapper(db),
		tableName: tableName,
		columns:   columns,
		metrics:   hook,
	}
}

// WithTx returns a lightweight, thread-safe clone of the repository bound to the provided transaction.
func (r *repository[T]) WithTx(tx Tx) Repository[T] {
	return &repository[T]{
		exec:      tx, // The clone now uses the transaction executor exclusively
		tableName: r.tableName,
		columns:   r.columns,
		metrics:   r.metrics,
	}
}

func (r *repository[T]) DB() Executor {
	return r.exec
}

func (r *repository[T]) GetQueryBuilder() squirrel.StatementBuilderType {
	return squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
}

func (r *repository[T]) Count(ctx context.Context, condition squirrel.Sqlizer) (int, error) {
	query := r.GetQueryBuilder().Select("count(*)").From(r.tableName)
	if condition != nil {
		query = query.Where(condition)
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return 0, WrapSQLError(err)
	}

	start := time.Now()
	var count int
	if err = r.exec.GetContext(ctx, &count, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, r.tableName, "count", start, err)
		}
		return 0, WrapSQLError(err)
	}

	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "count", start, nil)
	}
	return count, nil
}

func (r *repository[T]) Get(ctx context.Context, condition squirrel.Sqlizer, columns *[]string) (*T, error) {
	if condition == nil {
		return nil, ErrConditionRequired
	}
	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	sqlStr, args, err := r.GetQueryBuilder().Select(cols...).From(r.tableName).Where(condition).ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}

	start := time.Now()
	result := new(T)
	if err = r.exec.GetContext(ctx, result, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, r.tableName, "get", start, err)
		}
		return nil, WrapSQLError(err)
	}

	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "get", start, nil)
	}
	return result, nil
}

func (r *repository[T]) Select(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string) (*[]T, error) {
	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	query := r.GetQueryBuilder().Select(cols...).From(r.tableName)
	if condition != nil {
		query = query.Where(condition)
	}

	if sort != "" {
		orderBys, sortErr := r.buildSafeOrderBy(sort)
		if sortErr != nil {
			return nil, WrapSQLError(sortErr)
		}
		query = query.OrderBy(orderBys...)
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}

	start := time.Now()
	var results []T
	if err = r.exec.SelectContext(ctx, &results, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, r.tableName, "select", start, err)
		}
		return nil, WrapSQLError(err)
	}

	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "select", start, nil)
	}
	return &results, nil
}

func (r *repository[T]) PaginatedSelect(ctx context.Context, condition squirrel.Sqlizer, columns *[]string, sort string, limit, offset uint64) (*[]T, error) {
	if sort == "" {
		return nil, errors.New("repository: sort parameter is required for PaginatedSelect")
	}

	cols := r.columns
	if columns != nil {
		cols = *columns
	}

	query := r.GetQueryBuilder().Select(cols...).From(r.tableName)
	if condition != nil {
		query = query.Where(condition)
	}

	orderBys, sortErr := r.buildSafeOrderBy(sort)
	if sortErr != nil {
		return nil, WrapSQLError(sortErr)
	}
	query = query.OrderBy(orderBys...).Limit(limit).Offset(offset)

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return nil, WrapSQLError(err)
	}

	start := time.Now()
	var results []T
	if err = r.exec.SelectContext(ctx, &results, sqlStr, args...); err != nil {
		if r.metrics != nil {
			recordMetrics(r.metrics, r.tableName, "select", start, err)
		}
		return nil, WrapSQLError(err)
	}

	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "select", start, nil)
	}
	return &results, nil
}

func (r *repository[T]) Create(ctx context.Context, data map[string]any) error {
	if len(data) == 0 {
		return WrapSQLError(errors.New("repository: create data must not be empty"))
	}
	sqlStr, args, err := r.GetQueryBuilder().Insert(r.tableName).SetMap(data).ToSql()
	if err != nil {
		return WrapSQLError(err)
	}

	start := time.Now()
	_, err = r.exec.ExecContext(ctx, sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "exec", start, err)
	}

	return WrapSQLError(err)
}

func (r *repository[T]) CreateMany(ctx context.Context, columns []string, data []map[string]any, ignoreConflict bool) error {
	if len(data) == 0 || len(columns) == 0 {
		return nil
	}

	query := r.GetQueryBuilder().Insert(r.tableName).Columns(columns...)
	for _, row := range data {
		rowValues := make([]any, len(columns))
		for i, col := range columns {
			rowValues[i] = row[col]
		}
		query = query.Values(rowValues...)
	}
	if ignoreConflict {
		query = query.Suffix("ON CONFLICT DO NOTHING")
	}

	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}

	start := time.Now()
	_, err = r.exec.ExecContext(ctx, sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "exec", start, err)
	}

	return WrapSQLError(err)
}

func (r *repository[T]) Delete(ctx context.Context, condition squirrel.Sqlizer) error {
	if condition == nil {
		return ErrConditionRequired
	}

	query := r.GetQueryBuilder().Delete(r.tableName).Where(condition)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}

	start := time.Now()
	result, err := r.exec.ExecContext(ctx, sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "exec", start, err)
	}
	if err != nil {
		return WrapSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return WrapSQLError(err)
	}
	if rowsAffected == 0 {
		return ErrNotFoundNoRowsAffected()
	}

	return nil
}

func (r *repository[T]) Update(ctx context.Context, condition squirrel.Sqlizer, data map[string]any) error {
	if condition == nil {
		return ErrConditionRequired
	}
	if len(data) == 0 {
		return WrapSQLError(errors.New("repository: update data must not be empty"))
	}

	query := r.GetQueryBuilder().Update(r.tableName).SetMap(data).Where(condition)
	sqlStr, args, err := query.ToSql()
	if err != nil {
		return WrapSQLError(err)
	}

	start := time.Now()
	result, err := r.exec.ExecContext(ctx, sqlStr, args...)
	if r.metrics != nil {
		recordMetrics(r.metrics, r.tableName, "exec", start, err)
	}
	if err != nil {
		return WrapSQLError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return WrapSQLError(err)
	}
	if rowsAffected == 0 {
		return ErrNotFoundNoRowsAffected()
	}

	return nil
}

func (r *repository[T]) buildSafeOrderBy(sort string) ([]string, error) {
	allowedColumns := make(map[string]string, len(r.columns))
	for _, column := range r.columns {
		normalized := normalizeSortIdentifier(column)
		if normalized != "" {
			allowedColumns[normalized] = normalized
		}
	}

	parts := strings.Split(sort, ",")
	orderBys := make([]string, 0, len(parts))
	for _, part := range parts {
		clause := strings.TrimSpace(part)
		if clause == "" {
			continue
		}

		tokens := strings.Fields(clause)
		if len(tokens) == 0 || len(tokens) > 2 {
			return nil, fmt.Errorf("repository: invalid sort clause: %q", clause)
		}

		column := normalizeSortIdentifier(tokens[0])
		safeColumn, ok := allowedColumns[column]
		if !ok {
			return nil, fmt.Errorf("repository: invalid sort column: %q", tokens[0])
		}

		direction := "ASC"
		if len(tokens) == 2 {
			switch strings.ToUpper(tokens[1]) {
			case "ASC", "DESC":
				direction = strings.ToUpper(tokens[1])
			default:
				return nil, fmt.Errorf("repository: invalid sort direction: %q", tokens[1])
			}
		}
		orderBys = append(orderBys, fmt.Sprintf("%s %s", safeColumn, direction))
	}

	if len(orderBys) == 0 {
		return nil, errors.New("repository: sort must not be empty")
	}
	return orderBys, nil
}

func normalizeSortIdentifier(identifier string) string {
	trimmed := strings.TrimSpace(identifier)
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, ".")
	last := strings.TrimSpace(parts[len(parts)-1])
	last = strings.Trim(last, `"`)
	return strings.ToLower(last)
}
