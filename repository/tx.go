package yca_repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	yca_observer "github.com/yca-software/yca-go-core/observer"
)

// RunInTx begins a transaction on the provided DB pool, passes the wrapped Tx to the provided function,
// and manages the Commit/Rollback lifecycle automatically.
func RunInTx(ctx context.Context, db *sqlx.DB, hook yca_observer.QueryMetricsHook, fn func(tx Tx) error) error {
	rawTx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("repository: failed to begin transaction: %w", err)
	}

	tx := NewTxWrapper(rawTx)

	// Ensure rollback on panic or error
	defer func() {
		_ = tx.Rollback()
	}()

	// Execute business logic with the transaction
	if err := fn(tx); err != nil {
		return err // The defer will handle the rollback
	}

	// Commit if successful
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository: failed to commit transaction: %w", err)
	}

	return nil
}
