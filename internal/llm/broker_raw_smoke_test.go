//go:build smoke

package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Lower-level probe: see exactly what Go's HTTP stack reads off the broker.
// BROKER_SMOKE=1 BROKER_ENDPOINT=... BROKER_APIKEY=... \
//   go test -tags=smoke -run TestBrokerRawStream ./internal/llm -v -timeout 5m
func TestBrokerRawStream(t *testing.T) {
	if os.Getenv("BROKER_SMOKE") != "1" {
		t.Skip("set BROKER_SMOKE=1")
	}
	endpoint := strings.TrimRight(os.Getenv("BROKER_ENDPOINT"), "/")
	apiKey := os.Getenv("BROKER_APIKEY")
	model := os.Getenv("BROKER_MODEL")
	if model == "" {
		model = "qwen3:14b"
	}
	body, _ := json.Marshal(map[string]any{
		"model":    model,
		"stream":   true,
		"messages": []map[string]string{{"role": "user", "content": "say hi in one word"}},
	})

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "POST", endpoint+"/api/chat", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	defer resp.Body.Close()
	t.Logf("status=%d ct=%q broker=%q", resp.StatusCode, resp.Header.Get("Content-Type"), resp.Header.Get("X-Ollama-Broker"))

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	start := time.Now()
	lineN := 0
	for scanner.Scan() {
		lineN++
		line := scanner.Text()
		if line == "" {
			continue
		}
		t.Logf("[%6.2fs] line %d: %s", time.Since(start).Seconds(), lineN, truncate(line, 200))
		if lineN > 60 {
			t.Log("stopping after 60 lines")
			return
		}
	}
	if err := scanner.Err(); err != nil {
		t.Logf("scanner.Err: %v", err)
	}
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Logf("drain after EOF: %v", err)
	}
	t.Logf("stream ended after %d lines, %.2fs", lineN, time.Since(start).Seconds())
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
