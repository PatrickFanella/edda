package llm

import (
	"context"
	"errors"
	"log/slog"
)

// FallbackProvider sends each request to Primary and, when Primary returns an
// error, sends the same request to Secondary. A canceled or expired request
// context is returned as-is.
type FallbackProvider struct {
	Primary       Provider
	Secondary     Provider
	PrimaryName   string
	SecondaryName string
}

// Complete implements Provider.
func (f *FallbackProvider) Complete(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	resp, err := f.Primary.Complete(ctx, messages, tools)
	if err == nil || !f.shouldFallBack(ctx, err) {
		return resp, err
	}
	resp, err2 := f.Secondary.Complete(ctx, messages, tools)
	if err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return resp, nil
}

// Stream implements Provider. It falls back only when Primary fails before
// its stream opens.
func (f *FallbackProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	ch, err := f.Primary.Stream(ctx, messages, tools)
	if err == nil || !f.shouldFallBack(ctx, err) {
		return ch, err
	}
	ch, err2 := f.Secondary.Stream(ctx, messages, tools)
	if err2 != nil {
		return nil, errors.Join(err, err2)
	}
	return ch, nil
}

func (f *FallbackProvider) shouldFallBack(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	slog.Warn("llm primary provider failed; using fallback provider",
		"primary", f.PrimaryName, "fallback", f.SecondaryName, "error", err)
	return true
}
