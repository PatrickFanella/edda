//go:build integration

package handlers_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// Pause the second reservation immediately after source selection. A corrected
// finalizer attempts the same campaign lock; the old autocommit finalizer instead
// completes the source before the reservation's subsequent lease check.
type chapterRaceDB struct {
	*pgxpool.Pool
	release    chan struct{}
	finalizing chan struct{}
	finishOnce sync.Once
}
type chapterRaceTx struct {
	pgx.Tx
	owner *chapterRaceDB
}
type chapterRaceRow struct {
	pgx.Row
	scanned func()
}

func (r chapterRaceRow) Scan(dest ...any) error {
	err := r.Row.Scan(dest...)
	if err == nil && r.scanned != nil {
		r.scanned()
	}
	return err
}
func (d *chapterRaceDB) Begin(ctx context.Context) (pgx.Tx, error) {
	tx, err := d.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &chapterRaceTx{Tx: tx, owner: d}, nil
}
func (d *chapterRaceDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := d.Pool.QueryRow(ctx, sql, args...)
	if strings.HasPrefix(sql, "UPDATE chapter_revisions SET status=") {
		return chapterRaceRow{Row: row, scanned: func() { d.finishOnce.Do(func() { close(d.finalizing) }) }}
	}
	return row
}
func (tx *chapterRaceTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if strings.Contains(sql, "FOR UPDATE") && strings.Contains(sql, "JOIN chapter_revisions") {
		tx.owner.finishOnce.Do(func() { close(tx.owner.finalizing) })
	}
	row := tx.Tx.QueryRow(ctx, sql, args...)
	if strings.Contains(sql, "FROM chapters c WHERE campaign_id=$1 AND EXISTS") && strings.Contains(sql, "AND NOT EXISTS") {
		return chapterRaceRow{Row: row, scanned: func() {
			close(tx.owner.release)
			select {
			case <-tx.owner.finalizing:
			case <-ctx.Done():
			}
		}}
	}
	return row
}
func TestIntegrationChapterCompletionReservationRace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cid := uuid.New()
	if _, err := testPool.Exec(ctx, `INSERT INTO campaigns(id,name,status,created_by) VALUES($1,'Chapter race','active',$2)`, cid, testUserID); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := testPool.Exec(context.Background(), `DELETE FROM campaigns WHERE id=$1`, cid); err != nil {
			t.Error(err)
		}
	}()
	if _, err := testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response) VALUES($1,1,'wait','game_action','A quiet pause.')`, cid); err != nil {
		t.Fatal(err)
	}
	database := &chapterRaceDB{Pool: testPool, release: make(chan struct{}), finalizing: make(chan struct{})}
	entered := make(chan struct{})
	var enterOnce sync.Once
	model := &chapterModel{complete: func(context.Context) (*llm.Response, error) {
		enterOnce.Do(func() { close(entered) })
		select {
		case <-database.release:
		case <-ctx.Done():
		}
		return &llm.Response{Content: "A quiet pause."}, nil
	}}
	service := journal.NewSummarizer(model, journal.NewStore(database))
	type outcome struct {
		revision *api.ChapterRevisionResponse
		err      error
	}
	done := make(chan outcome, 1)
	go func() {
		revision, err := service.GenerateChapter(ctx, cid, uuid.New(), "")
		done <- outcome{revision, err}
	}()
	select {
	case <-entered:
	case initial := <-done:
		t.Fatalf("initial generation failed before model: %+v", initial)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	second, err := service.GenerateChapter(ctx, cid, uuid.New(), "")
	if err != journal.ErrChapterBusy {
		t.Errorf("next-range request unexpectedly reserved regeneration: %+v, %v", second, err)
	}
	first := <-done
	if first.err != nil || first.revision == nil || first.revision.Status != "completed" {
		t.Fatalf("first generation: %+v", first)
	}
	chapters, err := journal.NewStore(testPool).ListChapters(ctx, cid)
	if err != nil || len(chapters) != 1 || len(chapters[0].Revisions) != 1 {
		t.Fatalf("unrequested revision: %+v %v", chapters, err)
	}
}
