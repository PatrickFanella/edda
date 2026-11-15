package engine

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"git.subcult.tv/subculture-collective/edda/internal/journal"
)

type summaryContextDB struct{ contexts chan context.Context }

func (s *summaryContextDB) Query(ctx context.Context, _ string, _ ...any) (pgx.Rows, error) {
	s.contexts <- ctx
	return nil, errors.New("injected summary read failure")
}
func (*summaryContextDB) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}
func (*summaryContextDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func TestAutoSummaryLifetime(t *testing.T) {
	for _, useProcessContext := range []bool{false, true} {
		name := "default"
		if useProcessContext {
			name = "process"
		}
		t.Run(name, func(t *testing.T) {
			store := &summaryContextDB{contexts: make(chan context.Context, 1)}
			e := &Engine{logger: slog.New(slog.NewTextHandler(io.Discard, nil)), summarizer: journal.NewSummarizer(nil, journal.NewStore(store))}
			processCtx, stop := context.WithCancel(context.Background())
			defer stop()
			if useProcessContext {
				WithSummaryContext(processCtx)(e)
			}
			requestCtx, cancelRequest := context.WithCancel(context.Background())
			cancelRequest()
			for _, turn := range []int{-10, 0, 1, 9, 11, 19, 21} {
				e.autoSummarizeIfNeeded(requestCtx, uuid.New(), turn)
			}
			select {
			case <-store.contexts:
				t.Fatal("summary triggered outside a positive ten-turn boundary")
			default:
			}
			started := time.Now()
			e.autoSummarizeIfNeeded(requestCtx, uuid.New(), 10)
			select {
			case summaryCtx := <-store.contexts:
				deadline, ok := summaryCtx.Deadline()
				if !ok || deadline.Before(started.Add(119*time.Second)) || deadline.After(started.Add(121*time.Second)) {
					t.Fatalf("summary deadline = %v (present=%v)", deadline, ok)
				}
			case <-time.After(time.Second):
				t.Fatal("summary did not start")
			}
		})
	}
}
