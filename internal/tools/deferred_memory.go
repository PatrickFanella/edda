package tools

import (
	"context"
	"log/slog"
)

type deferredMemoryKey struct{}
type memoryCallbacks struct{ jobs []func(context.Context) error }

// WithDeferredMemory collects auxiliary embeddings while a mechanical tool runs.
// Flush only after its database commit, using a context without that transaction.
func WithDeferredMemory(ctx context.Context) (context.Context, func(context.Context)) {
	callbacks := &memoryCallbacks{}
	return context.WithValue(ctx, deferredMemoryKey{}, callbacks), func(base context.Context) {
		for _, job := range callbacks.jobs {
			if err := job(base); err != nil {
				slog.Warn("post-commit memory update failed", "error", err)
			}
		}
	}
}
func runMemoryEffect(ctx context.Context, job func(context.Context) error) error {
	if callbacks, ok := ctx.Value(deferredMemoryKey{}).(*memoryCallbacks); ok {
		callbacks.jobs = append(callbacks.jobs, job)
		return nil
	}
	return job(ctx)
}
