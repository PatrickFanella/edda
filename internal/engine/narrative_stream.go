package engine

import (
	"context"
	"io"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/llm"
)

// completeNarrative uses provider fragments only for a live caller. Tool deltas
// are complete calls at this boundary and are never dispatched before Done.
// Empty callback text resets the preview; it never changes durable history.
func completeNarrative(ctx context.Context, provider llm.Provider, messages []llm.Message, tools []llm.Tool, emit func(string)) (resp *llm.Response, err error) {
	if emit == nil {
		return provider.Complete(ctx, messages, tools)
	}
	emit("")
	defer func() {
		if err != nil {
			emit("")
		}
	}()
	streamCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	chunks, err := provider.Stream(streamCtx, messages, tools)
	if err != nil {
		return nil, err
	}
	var text strings.Builder
	var calls []llm.ToolCall
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case chunk, ok := <-chunks:
			if !ok {
				return nil, &llm.ErrTransient{Err: io.ErrUnexpectedEOF}
			}
			if chunk.Reset {
				text.Reset()
				calls = nil
				emit("")
			}
			if chunk.ContentDelta != "" {
				text.WriteString(chunk.ContentDelta)
				emit(chunk.ContentDelta)
			}
			if chunk.ToolCallDelta != nil {
				calls = append(calls, *chunk.ToolCallDelta)
			}
			if chunk.Done {
				return &llm.Response{Content: text.String(), ToolCalls: calls}, nil
			}
		}
	}
}
