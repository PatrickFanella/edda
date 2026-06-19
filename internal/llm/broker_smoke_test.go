//go:build smoke

package llm

import (
	"context"
	"os"
	"testing"
	"time"
)

// Run with: BROKER_SMOKE=1 BROKER_ENDPOINT=http://10.0.0.50:11434 BROKER_APIKEY=... \
//   go test -tags=smoke -run TestBrokerSmoke ./internal/llm -v
func TestBrokerSmoke(t *testing.T) {
	if os.Getenv("BROKER_SMOKE") != "1" {
		t.Skip("set BROKER_SMOKE=1 to run live broker smoke")
	}
	endpoint := os.Getenv("BROKER_ENDPOINT")
	apiKey := os.Getenv("BROKER_APIKEY")
	model := os.Getenv("BROKER_MODEL")
	if model == "" {
		model = "qwen3:14b"
	}
	if endpoint == "" || apiKey == "" {
		t.Fatal("BROKER_ENDPOINT and BROKER_APIKEY required")
	}

	client := NewOllamaClient(endpoint, model).WithAPIKey(apiKey)
	msgs := []Message{{Role: RoleUser, Content: "Reply with exactly one short word: pong"}}

	t.Run("Complete-SSE-unwrap", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		resp, err := client.Complete(ctx, msgs, nil)
		if err != nil {
			t.Fatalf("Complete() error: %v", err)
		}
		if resp == nil || resp.Content == "" {
			t.Fatalf("empty content; full resp: %+v", resp)
		}
		t.Logf("Complete: content=%q usage=%+v finish=%q",
			resp.Content, resp.Usage, resp.FinishReason)
	})

	t.Run("Stream-SSE-with-DONE", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
		defer cancel()
		ch, err := client.Stream(ctx, []Message{{Role: RoleUser, Content: "Count from 1 to 5, comma-separated, then stop."}}, nil)
		if err != nil {
			t.Fatalf("Stream() error: %v", err)
		}
		var chunks, doneCount int
		var full string
		for chunk := range ch {
			chunks++
			full += chunk.ContentDelta
			if chunk.Done {
				doneCount++
			}
		}
		if chunks == 0 || full == "" {
			t.Fatalf("no chunks received (chunks=%d, full=%q)", chunks, full)
		}
		if doneCount != 1 {
			t.Errorf("expected exactly 1 Done chunk, got %d", doneCount)
		}
		t.Logf("Stream: chunks=%d done=%d total=%q", chunks, doneCount, full)
	})
}
