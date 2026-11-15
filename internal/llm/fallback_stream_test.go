package llm

import (
	"context"
	"sync/atomic"
	"testing"
)

type fallbackStreamFixture struct {
	Provider
	chunks []StreamChunk
	calls  atomic.Int32
}

func (p *fallbackStreamFixture) Stream(context.Context, []Message, []Tool) (<-chan StreamChunk, error) {
	p.calls.Add(1)
	ch := make(chan StreamChunk, len(p.chunks))
	for _, c := range p.chunks {
		ch <- c
	}
	close(ch)
	return ch, nil
}
func TestFallbackStreamRecoversOpenedFailure(t *testing.T) {
	primary := &fallbackStreamFixture{chunks: []StreamChunk{{ContentDelta: "Abandoned"}, {ToolCallDelta: &ToolCall{ID: "old", Name: "update_player_hp"}}}}
	secondary := &fallbackStreamFixture{chunks: []StreamChunk{{ContentDelta: "Recovered"}, {Done: true}}}
	fallback := &FallbackProvider{Primary: primary, Secondary: secondary}
	ch, err := fallback.Stream(context.Background(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var chunks []StreamChunk
	for chunk := range ch {
		chunks = append(chunks, chunk)
	}
	if secondary.calls.Load() != 1 {
		t.Fatalf("secondary calls=%d", secondary.calls.Load())
	}
	if len(chunks) != 5 || !chunks[2].Reset || chunks[3].ContentDelta != "Recovered" || !chunks[4].Done {
		t.Fatalf("fallback did not replace abandoned attempt: %+v", chunks)
	}
}
func TestFallbackStreamDoesNotSwitchAfterCompletionOrCancellation(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(map[bool]string{true: "canceled", false: "completed"}[canceled], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if canceled {
				cancel()
			}
			primary := &fallbackStreamFixture{chunks: []StreamChunk{{ContentDelta: "Done"}, {Done: true}}}
			secondary := &fallbackStreamFixture{chunks: []StreamChunk{{Done: true}}}
			ch, err := (&FallbackProvider{Primary: primary, Secondary: secondary}).Stream(ctx, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			for range ch {
			}
			if secondary.calls.Load() != 0 {
				t.Fatal("unexpected fallback")
			}
		})
	}
}
