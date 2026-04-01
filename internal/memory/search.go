package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	pgvector_go "github.com/pgvector/pgvector-go"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

// MemorySearchStore is the narrow persistence interface required by Searcher.
// It is satisfied by statedb.Queries.
type MemorySearchStore interface {
	SearchMemoriesBySimilarity(ctx context.Context, arg statedb.SearchMemoriesBySimilarityParams) ([]statedb.SearchMemoriesBySimilarityRow, error)
}

// MemoryResult is a domain-level representation of a memory returned by
// similarity search. Distance is cosine distance: lower values indicate
// higher similarity.
type MemoryResult struct {
	ID         uuid.UUID
	CampaignID uuid.UUID
	Content    string
	MemoryType string
	Distance   float64   // cosine distance (lower = more similar)
	CreatedAt  time.Time
}

// Searcher performs semantic similarity search over memories using a vector
// embedding of the query text.
type Searcher struct {
	embedder Embedder
	store    MemorySearchStore
}

// NewSearcher constructs a Searcher backed by the given embedder and store.
func NewSearcher(embedder Embedder, store MemorySearchStore) *Searcher {
	return &Searcher{embedder: embedder, store: store}
}

// defaultSearchLimit is used when the caller supplies a non-positive limit.
const defaultSearchLimit = 5

// SearchSimilar returns memories most similar to query for the given campaign,
// ordered by ascending cosine distance (most similar first).
func (s *Searcher) SearchSimilar(ctx context.Context, campaignID uuid.UUID, query string, limit int) ([]MemoryResult, error) {
	if query == "" {
		return nil, fmt.Errorf("memory search: %w", &ErrEmptyInput{})
	}
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	vec, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("memory search: embed query: %w", err)
	}

	params := statedb.SearchMemoriesBySimilarityParams{
		QueryEmbedding: pgvector_go.NewVector(vec),
		CampaignID:     dbutil.ToPgtype(campaignID),
		LimitCount:     int32(limit),
	}

	rows, err := s.store.SearchMemoriesBySimilarity(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("memory search: query store: %w", err)
	}

	results := make([]MemoryResult, len(rows))
	for i, r := range rows {
		results[i] = MemoryResult{
			ID:         dbutil.FromPgtype(r.ID),
			CampaignID: dbutil.FromPgtype(r.CampaignID),
			Content:    r.Content,
			MemoryType: r.MemoryType,
			Distance:   r.Distance,
			CreatedAt:  r.CreatedAt.Time,
		}
	}
	return results, nil
}
