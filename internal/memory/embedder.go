// Package memory defines the provider-agnostic embedding interface and
// supporting types used by the game-master memory subsystem. Implementations
// may wrap any embedding provider (e.g. Ollama with nomic-embed-text, OpenAI,
// or a local model) while callers depend only on this interface.
package memory

import (
	"context"
	"time"
)

// Embedder generates vector embeddings from text. Implementations must be safe
// for concurrent use and should respect the supplied context for cancellation
// and deadline propagation.
//
// All returned vectors must have length equal to the configured vector
// dimension (see DefaultVectorDimension). Implementations should return an
// ErrDimensionMismatch if the underlying model produces vectors of an
// unexpected size.
type Embedder interface {
	// Embed returns a single vector embedding for the given text.
	// The returned slice must have length equal to the configured vector
	// dimension. An error is returned if the text is empty or if the
	// provider fails to produce an embedding.
	Embed(ctx context.Context, text string) ([]float32, error)

	// BatchEmbed returns vector embeddings for multiple texts in a single
	// call, improving throughput when many embeddings are needed at once.
	// The returned outer slice has the same length and order as texts.
	// An error is returned if texts is empty, if any individual text is
	// empty, or if the provider fails.
	BatchEmbed(ctx context.Context, texts []string) ([][]float32, error)
}

// EmbeddingResult pairs a computed vector with metadata about the embedding
// operation. It is intended for callers that need to inspect or log details
// beyond the raw vector.
type EmbeddingResult struct {
	// Vector is the computed embedding. Its length equals the configured
	// vector dimension.
	Vector []float32

	// Text is the original input that was embedded.
	Text string

	// Model identifies the embedding model that produced this vector
	// (e.g. "nomic-embed-text").
	Model string

	// Dimension is the length of Vector.
	Dimension int

	// Duration is the wall-clock time spent producing this embedding.
	Duration time.Duration
}
