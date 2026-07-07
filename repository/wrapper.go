package chi_repository

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// --- DB Wrapper (Connection Pool) ---

type DBWrapper struct {
	*sqlx.DB
}

func NewDBWrapper(db *sqlx.DB) *DBWrapper {
	if db == nil {
		panic("repository: db must not be nil")
	}
	return &DBWrapper{DB: db}
}

func (w *DBWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.DB.ExecContext(ctx, query, args...)
}

func (w *DBWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return w.DB.GetContext(ctx, dest, query, args...)
}

func (w *DBWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return w.DB.SelectContext(ctx, dest, query, args...)
}

// --- Tx Wrapper (Transaction) ---

type TxWrapper struct {
	*sqlx.Tx
}

func NewTxWrapper(tx *sqlx.Tx) *TxWrapper {
	if tx == nil {
		panic("repository: tx must not be nil")
	}
	return &TxWrapper{Tx: tx}
}

func (w *TxWrapper) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return w.Tx.ExecContext(ctx, query, args...)
}

func (w *TxWrapper) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return w.Tx.GetContext(ctx, dest, query, args...)
}

func (w *TxWrapper) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return w.Tx.SelectContext(ctx, dest, query, args...)
}
