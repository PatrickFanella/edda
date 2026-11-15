package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (e *Engine) acquireTurn(ctx context.Context, id uuid.UUID) (func(), error) {
	if e.turnLock != nil {
		return e.turnLock(ctx, id)
	}
	return localTurnLock(ctx, id)
}

// dedicatedTurnLock uses a dedicated connection rather than holding a query-pool
// slot across inference. Otherwise concurrent campaigns could exhaust the pool
// with lock holders that are all waiting for a connection to load/save state.
func dedicatedTurnLock(cfg *pgx.ConnConfig) turnLock {
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

		var acquired bool
		err = conn.QueryRow(ctx, `SELECT pg_try_advisory_lock(hashtextextended('edda:turn:' || $1::text, 0))`, id.String()).Scan(&acquired)
		if err != nil {
			closeConn()
			return nil, fmt.Errorf("acquire turn lock: %w", err)
		}
		if !acquired {
			closeConn()
			return nil, ErrTurnInProgress
		}
		// Closing this dedicated connection releases its session lock. No open
		// transaction or query-pool slot survives inference.
		return closeConn, nil
	}
}
