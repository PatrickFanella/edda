package memory

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const validateMemoryEmbeddingDimensionSQL = `SELECT atttypmod::int FROM pg_attribute WHERE attrelid = 'memories'::regclass AND attname = 'embedding'`

var ErrMemoryEmbeddingDimensionMismatch = errors.New("memory embedding dimension mismatch")

type memoryDimensionQueryRow interface {
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// ValidateMemoryEmbeddingDimension checks that the persisted memories.embedding
// vector dimension matches the configured embedding dimension.
func ValidateMemoryEmbeddingDimension(ctx context.Context, db memoryDimensionQueryRow, configuredDimension int) error {
	var dbDimension int
	if err := db.QueryRow(ctx, validateMemoryEmbeddingDimensionSQL).Scan(&dbDimension); err != nil {
		return fmt.Errorf("validate memory embedding dimension: %w", err)
	}
	if dbDimension != configuredDimension {
		return fmt.Errorf("%w: db=%d configured=%d", ErrMemoryEmbeddingDimensionMismatch, dbDimension, configuredDimension)
	}
	return nil
}
