package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrTurnInProgress tells clients to retry after the active campaign turn ends.
var ErrTurnInProgress = errors.New("a turn is already in progress for this campaign; retry after it finishes")

type transactionBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

type turnLock func(context.Context, uuid.UUID) (func(), error)

// Used by in-memory engines in tests. Database-backed engines always use the
// PostgreSQL lock so independent processes share the same exclusion boundary.
var localTurns sync.Map

func localTurnLock(ctx context.Context, id uuid.UUID) (func(), error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if _, loaded := localTurns.LoadOrStore(id, struct{}{}); loaded {
		return nil, ErrTurnInProgress
	}
	return func() { localTurns.Delete(id) }, nil
}

func postgresTurnLock(db transactionBeginner) turnLock {
	return func(ctx context.Context, id uuid.UUID) (func(), error) {
		tx, err := db.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("begin turn lock: %w", err)
		}
		release := func() {
			// Request cancellation must not strand a lock/connection.
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			// Rollback releases the transaction advisory lock. pgx discards the
			// connection if rollback fails, releasing its locks server-side as well.
			_ = tx.Rollback(cleanup)
		}
		var acquired bool
		err = tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended('edda:turn:' || $1::text, 0))`, id.String()).Scan(&acquired)
		if err != nil {
			release()
			return nil, fmt.Errorf("acquire turn lock: %w", err)
		}
		if !acquired {
			release()
			return nil, ErrTurnInProgress
		}
		return release, nil
	}
}

func (e *Engine) acquireTurn(ctx context.Context, id uuid.UUID) (func(), error) {
	if e.turnLock != nil {
		return e.turnLock(ctx, id)
	}
	return localTurnLock(ctx, id)
}

// poolTurnLock uses a dedicated connection rather than holding a query-pool
// slot across inference. Otherwise concurrent campaigns could exhaust the pool
// with lock holders that are all waiting for a connection to load/save state.
func poolTurnLock(pool *pgxpool.Pool) turnLock {
	cfg := pool.Config().ConnConfig.Copy()
	return func(ctx context.Context, id uuid.UUID) (func(), error) {
		conn, err := pgx.ConnectConfig(ctx, cfg.Copy())
		if err != nil {
			return nil, fmt.Errorf("connect turn lock: %w", err)
		}
		closeConn := func() {
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = conn.Close(cleanup)
		}
		release, err := postgresTurnLock(conn)(ctx, id)
		if err != nil {
			closeConn()
			return nil, err
		}
		return func() { release(); closeConn() }, nil
	}
}
