package engine

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
		// Wait for PostgreSQL to acknowledge unlock before returning a completed
		// turn. Closing the socket alone does not wait for backend disconnect
		// cleanup, so an immediate next request could otherwise see a stale lock.
		return func() {
			defer closeConn()
			cleanup, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			var unlocked bool
			if err := conn.QueryRow(cleanup, `SELECT pg_advisory_unlock(hashtextextended('edda:turn:' || $1::text, 0))`, id.String()).Scan(&unlocked); err != nil || !unlocked {
				slog.Warn("turn lock release was not acknowledged; closing connection", "campaign_id", id, "error", err)
			}
		}, nil
	}
}
