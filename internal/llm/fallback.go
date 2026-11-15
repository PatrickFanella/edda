package llm

import (
	"context"
	"errors"
	"io"
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

// Stream implements Provider. An opened primary stream that fails without Done
// is abandoned before trying Secondary. Reset invalidates its provisional text
// and calls; consumers must not dispatch tool calls until successful completion.
func (f *FallbackProvider) Stream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamChunk, error) {
	primaryCtx, cancelPrimary := context.WithCancel(ctx)
	primary, err := f.Primary.Stream(primaryCtx, messages, tools)
	if err != nil {
		cancelPrimary()
		if !f.shouldFallBack(ctx, err) {
			return nil, err
		}
		secondary, secondaryErr := f.Secondary.Stream(ctx, messages, tools)
		if secondaryErr != nil {
			return nil, errors.Join(err, secondaryErr)
		}
		return secondary, nil
	}
	out := make(chan StreamChunk)
	go func() {
		defer close(out)
		defer cancelPrimary()
		forward := func(source <-chan StreamChunk) bool {
			for {
				select {
				case <-ctx.Done():
					return false
				case chunk, ok := <-source:
					if !ok {
						return false
					}
					select {
					case out <- chunk:
					case <-ctx.Done():
						return false
					}
					if chunk.Done {
						return true
					}
				}
			}
		}
		if forward(primary) {
			return
		}
		cancelPrimary()
		if !f.shouldFallBack(ctx, io.ErrUnexpectedEOF) {
			return
		}
		select {
		case out <- StreamChunk{Reset: true}:
		case <-ctx.Done():
			return
		}
		secondaryCtx, cancelSecondary := context.WithCancel(ctx)
		defer cancelSecondary()
		secondary, err := f.Secondary.Stream(secondaryCtx, messages, tools)
		if err != nil {
			return
		}
		forward(secondary)
	}()
	return out, nil
}

func (f *FallbackProvider) shouldFallBack(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return false
	}
	slog.Warn("llm primary provider failed; using fallback provider",
		"primary", f.PrimaryName, "fallback", f.SecondaryName, "error", err)
	return true
}
