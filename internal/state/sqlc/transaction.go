package statedb

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// RunInTransaction binds all callback queries and raw writes to one transaction.
// The callback must not call external services.
func (q *Queries) RunInTransaction(ctx context.Context, fn func(Querier, DBTX) error) error {
	beginner, ok := q.db.(interface {
		Begin(context.Context) (pgx.Tx, error)
	})
	if !ok {
		return fmt.Errorf("database does not support transactions")
	}
	tx, err := beginner.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = tx.Rollback(rollbackCtx)
	}()
	if err := fn(q.WithTx(tx), tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
