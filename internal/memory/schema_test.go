package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeMemoryDimensionDB struct {
	dimension int
	err       error
}

type fakeMemoryDimensionRow struct {
	dimension int
	err       error
}

func (f fakeMemoryDimensionDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return fakeMemoryDimensionRow{dimension: f.dimension, err: f.err}
}

func (r fakeMemoryDimensionRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != 1 {
		return errors.New("unexpected dest count")
	}
	*(dest[0].(*int)) = r.dimension
	return nil
}

func TestValidateMemoryEmbeddingDimension(t *testing.T) {
	if err := ValidateMemoryEmbeddingDimension(context.Background(), fakeMemoryDimensionDB{dimension: 768}, 768); err != nil {
		t.Fatalf("ValidateMemoryEmbeddingDimension() error = %v", err)
	}
}

func TestValidateMemoryEmbeddingDimensionMismatch(t *testing.T) {
	err := ValidateMemoryEmbeddingDimension(context.Background(), fakeMemoryDimensionDB{dimension: 1536}, 768)
	if err == nil {
		t.Fatal("expected mismatch error, got nil")
	}
	if !errors.Is(err, ErrMemoryEmbeddingDimensionMismatch) {
		t.Fatalf("expected ErrMemoryEmbeddingDimensionMismatch, got %v", err)
	}
}
