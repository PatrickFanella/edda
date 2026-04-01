package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	pgvector_go "github.com/pgvector/pgvector-go"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// --- stubs ---

type stubSearchEmbedder struct {
	err error
}

func (s *stubSearchEmbedder) Embed(_ context.Context, _ string) ([]float32, error) {
	if s.err != nil {
		return nil, s.err
	}
	return make([]float32, DefaultVectorDimension), nil
}

func (s *stubSearchEmbedder) BatchEmbed(_ context.Context, texts []string) ([][]float32, error) {
	if s.err != nil {
		return nil, s.err
	}
	out := make([][]float32, len(texts))
	for i := range texts {
		out[i] = make([]float32, DefaultVectorDimension)
	}
	return out, nil
}

type stubMemorySearchStore struct {
	params statedb.SearchMemoriesBySimilarityParams
	rows   []statedb.SearchMemoriesBySimilarityRow
	err    error
}

func (s *stubMemorySearchStore) SearchMemoriesBySimilarity(_ context.Context, arg statedb.SearchMemoriesBySimilarityParams) ([]statedb.SearchMemoriesBySimilarityRow, error) {
	s.params = arg
	if s.err != nil {
		return nil, s.err
	}
	return s.rows, nil
}

// --- helpers ---

func makeRow(id, campID uuid.UUID, content, mtype string, dist float64, ts time.Time) statedb.SearchMemoriesBySimilarityRow {
	return statedb.SearchMemoriesBySimilarityRow{
		ID:         dbutil.ToPgtype(id),
		CampaignID: dbutil.ToPgtype(campID),
		Content:    content,
		MemoryType: mtype,
		Embedding:  pgvector_go.NewVector(make([]float32, DefaultVectorDimension)),
		LocationID: pgtype.UUID{},
		InGameTime: pgtype.Text{},
		CreatedAt:  pgtype.Timestamptz{Time: ts, Valid: true},
		Distance:   dist,
	}
}

// --- tests ---

func TestSearchSimilar_Success(t *testing.T) {
	campID := uuid.New()
	id1, id2 := uuid.New(), uuid.New()
	now := time.Now().Truncate(time.Microsecond)

	store := &stubMemorySearchStore{
		rows: []statedb.SearchMemoriesBySimilarityRow{
			makeRow(id1, campID, "battle at dawn", "event", 0.1, now),
			makeRow(id2, campID, "meeting the elder", "dialogue", 0.3, now.Add(-time.Hour)),
		},
	}
	s := NewSearcher(&stubSearchEmbedder{}, store)

	results, err := s.SearchSimilar(context.Background(), campID, "fight", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != id1 {
		t.Errorf("results[0].ID = %v, want %v", results[0].ID, id1)
	}
	if results[0].Content != "battle at dawn" {
		t.Errorf("results[0].Content = %q, want %q", results[0].Content, "battle at dawn")
	}
	if results[0].Distance != 0.1 {
		t.Errorf("results[0].Distance = %f, want 0.1", results[0].Distance)
	}
	if !results[0].CreatedAt.Equal(now) {
		t.Errorf("results[0].CreatedAt = %v, want %v", results[0].CreatedAt, now)
	}
	if results[1].MemoryType != "dialogue" {
		t.Errorf("results[1].MemoryType = %q, want %q", results[1].MemoryType, "dialogue")
	}
	// Verify params passed to store.
	if store.params.LimitCount != 10 {
		t.Errorf("store LimitCount = %d, want 10", store.params.LimitCount)
	}
}

func TestSearchSimilar_EmptyQuery(t *testing.T) {
	s := NewSearcher(&stubSearchEmbedder{}, &stubMemorySearchStore{})
	_, err := s.SearchSimilar(context.Background(), uuid.New(), "", 5)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
	var emptyErr *ErrEmptyInput
	if !errors.As(err, &emptyErr) {
		t.Errorf("expected ErrEmptyInput, got %T: %v", err, err)
	}
}

func TestSearchSimilar_EmbedError(t *testing.T) {
	embedErr := errors.New("provider down")
	s := NewSearcher(&stubSearchEmbedder{err: embedErr}, &stubMemorySearchStore{})
	_, err := s.SearchSimilar(context.Background(), uuid.New(), "hello", 5)
	if err == nil {
		t.Fatal("expected error from embedder")
	}
	if !errors.Is(err, embedErr) {
		t.Errorf("expected wrapped embedErr, got %v", err)
	}
}

func TestSearchSimilar_StoreError(t *testing.T) {
	storeErr := errors.New("db timeout")
	store := &stubMemorySearchStore{err: storeErr}
	s := NewSearcher(&stubSearchEmbedder{}, store)
	_, err := s.SearchSimilar(context.Background(), uuid.New(), "hello", 5)
	if err == nil {
		t.Fatal("expected error from store")
	}
	if !errors.Is(err, storeErr) {
		t.Errorf("expected wrapped storeErr, got %v", err)
	}
}

func TestSearchSimilar_DefaultLimit(t *testing.T) {
	store := &stubMemorySearchStore{}
	s := NewSearcher(&stubSearchEmbedder{}, store)

	// limit = 0 should default to 5
	_, err := s.SearchSimilar(context.Background(), uuid.New(), "query", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.params.LimitCount != 5 {
		t.Errorf("LimitCount = %d, want 5 (default)", store.params.LimitCount)
	}

	// limit = -1 should also default to 5
	_, err = s.SearchSimilar(context.Background(), uuid.New(), "query", -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.params.LimitCount != 5 {
		t.Errorf("LimitCount = %d, want 5 (default)", store.params.LimitCount)
	}
}

func TestSearchSimilar_EmptyResults(t *testing.T) {
	store := &stubMemorySearchStore{
		rows: []statedb.SearchMemoriesBySimilarityRow{}, // explicitly empty
	}
	s := NewSearcher(&stubSearchEmbedder{}, store)

	results, err := s.SearchSimilar(context.Background(), uuid.New(), "obscure", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Fatal("expected non-nil empty slice, got nil")
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
