package engine

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/assembly"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
)

type streamingState struct {
	stubStateManager
	mu   sync.Mutex
	logs []domain.SessionLog
	fail bool
}

func (s *streamingState) SaveSessionLog(_ context.Context, log domain.SessionLog) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.fail {
		return errors.New("fixture persistence failure")
	}
	s.logs = append(s.logs, log)
	return nil
}

func (s *streamingState) saved() []domain.SessionLog {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.SessionLog(nil), s.logs...)
}

// The real Ollama adapter must deliver each flushed fragment while the fixture
// withholds both the next fragment and the provider completion marker.
func TestProcessTurnStream_ProviderFragmentsPrecedePersistence(t *testing.T) {
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "persistence failure"}[fail], func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for _, text := range []string{"A quiet ", "room."} {
					if err := json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"content": text}, "done": false}); err != nil {
						return
					}
					w.(http.Flusher).Flush()
					select {
					case <-release:
					case <-r.Context().Done():
						return
					}
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"message": map[string]string{"content": ""}, "done": true}); err != nil {
					return
				}
			}))
			defer server.Close()
			defer cancel() // Close a paused request before shutting down the server.
			reg := tools.NewRegistry()
			provider := llm.NewOllamaClient(server.URL, "fixture")
			state := &streamingState{fail: fail}
			e := &Engine{logger: slog.Default(), state: state, assembler: assembly.NewContextAssembler(reg), processor: NewTurnProcessor(provider, reg, tools.NewValidator(reg), nil)}
			events, err := e.ProcessTurnStream(ctx, uuid.New(), "look")
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"A quiet ", "room."} {
				for {
					select {
					case event, ok := <-events:
						if !ok {
							t.Fatal("stream closed before fragments")
						}
						if event.Type == "result" || event.Type == "error" {
							t.Fatalf("premature terminal: %+v", event)
						}
						if event.Type != "chunk" {
							continue
						}
						if event.Text != want {
							t.Fatalf("fragment = %q, want %q", event.Text, want)
						}
						if len(state.saved()) != 0 {
							t.Fatal("saved before provider completed")
						}
					case <-ctx.Done():
						t.Fatal("fragment was buffered until provider completion")
					}
					break
				}
				release <- struct{}{}
			}
			results, failures := 0, 0
			for event := range events {
				switch event.Type {
				case "chunk":
					t.Fatal("full narrative was emitted again after persistence")
				case "result":
					results++
					if event.Result.Narrative != "A quiet room." || len(state.saved()) != 1 {
						t.Fatal("result did not follow exact persistence")
					}
				case "error":
					failures++
				}
			}
			if fail && (results != 0 || failures != 1 || len(state.saved()) != 0) {
				t.Fatal("failed persistence published a committed result")
			}
			if !fail && (results != 1 || failures != 0) {
				t.Fatal("missing successful result")
			}
		})
	}
}

type fragmentProvider struct {
	mockProvider
	streams     [][]llm.StreamChunk
	streamCalls int
}

func (p *fragmentProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	if p.streamCalls >= len(p.streams) {
		p.t.Fatal("unexpected stream request")
	}
	chunks := p.streams[p.streamCalls]
	p.streamCalls++
	ch := make(chan llm.StreamChunk, len(chunks))
	for _, chunk := range chunks {
		ch <- chunk
	}
	close(ch)
	return ch, nil
}

func TestTurnProcessor_ProvisionalRetryAndToolCorrection(t *testing.T) {
	for _, correction := range []bool{false, true} {
		t.Run(map[bool]string{false: "interrupted provider retry", true: "tool correction"}[correction], func(t *testing.T) {
			reg := tools.NewRegistry()
			p := &fragmentProvider{mockProvider: mockProvider{t: t}}
			if correction {
				// No movement tool can back this claim. The existing repair path
				// must clear it and stream the safe rewrite instead.
				p.streams = [][]llm.StreamChunk{
					{{ContentDelta: "You enter the castle."}, {Done: true}},
					{{ContentDelta: "The castle remains "}, {ContentDelta: "out of reach."}, {Done: true}},
				}
			} else {
				p.streams = [][]llm.StreamChunk{
					{{ContentDelta: "Abandoned text"}, {ToolCallDelta: &llm.ToolCall{Name: "probe"}}},
					{{ContentDelta: "A quiet "}, {ContentDelta: "room."}, {Done: true}},
				}
			}
			var preview string
			var discarded []string
			emit := func(text string) {
				if text == "" {
					discarded = append(discarded, preview)
					preview = ""
				} else {
					preview += text
				}
			}
			tp := NewTurnProcessor(p, reg, tools.NewValidator(reg), nil)
			narrative, applied, err := tp.ProcessWithRecoveryWithOptions(context.Background(), nil, nil, TurnProcessorOptions{NarrativeCallback: emit})
			if err != nil {
				t.Fatal(err)
			}
			if p.streamCalls != 2 || len(applied) != 0 || preview != narrative {
				t.Fatalf("unexpected result: %q, preview %q, calls %d", narrative, preview, p.streamCalls)
			}
			wantDiscarded := "Abandoned text"
			if correction {
				wantDiscarded = "You enter the castle."
			}
			if !strings.Contains(strings.Join(discarded, "|"), wantDiscarded) {
				t.Fatal("superseded preview was not cleared")
			}
		})
	}
}

func TestCompleteNarrative_CancelSilentProvider(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &silentFragmentProvider{opened: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := completeNarrative(ctx, p, nil, nil, func(string) {}); done <- err }()
	<-p.opened
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("silent provider retained engine consumer after cancellation")
	}
}

func TestTurnProcessor_StreamedToolsClearPreviewBeforeExecution(t *testing.T) {
	for _, retry := range []bool{false, true} {
		t.Run(map[bool]string{false: "tool-only continuation", true: "validation retry"}[retry], func(t *testing.T) {
			var preview string
			var fragments []string
			var executed int
			reg := tools.NewRegistry()
			tool := llm.Tool{Name: "probe", Parameters: map[string]any{"type": "object", "properties": map[string]any{"name": map[string]any{"type": "string"}}, "required": []any{"name"}}}
			if err := reg.Register(tool, func(context.Context, map[string]any) (*tools.ToolResult, error) {
				if preview != "" {
					t.Fatalf("tool executed while superseded preview %q remained", preview)
				}
				executed++
				return &tools.ToolResult{Success: true}, nil
			}); err != nil {
				t.Fatal(err)
			}
			call := llm.ToolCall{ID: "first", Name: "probe", Arguments: map[string]any{"name": "gate"}}
			p := &fragmentProvider{mockProvider: mockProvider{t: t}}
			if retry {
				p.responses = []*llm.Response{{ToolCalls: []llm.ToolCall{call}}}
				p.errors = []error{nil}
				call.Arguments = nil
				p.streams = [][]llm.StreamChunk{{{ContentDelta: "A quiet room."}, {ToolCallDelta: &call}, {Done: true}}}
			} else {
				p.streams = [][]llm.StreamChunk{
					{{ToolCallDelta: &call}, {Done: true}},
					{{ContentDelta: "A quiet "}, {ContentDelta: "room."}, {Done: true}},
				}
			}
			tp := NewTurnProcessor(p, reg, tools.NewValidator(reg), nil)
			text, applied, err := tp.ProcessWithRecoveryWithOptions(context.Background(), nil, []llm.Tool{tool}, TurnProcessorOptions{NarrativeCallback: func(text string) {
				if text == "" {
					preview = ""
				} else {
					preview += text
					fragments = append(fragments, text)
				}
			}})
			if err != nil || text != "A quiet room." || len(applied) != 1 || executed != 1 {
				t.Fatalf("text=%q, applied=%d, executed=%d, error=%v", text, len(applied), executed, err)
			}
			if retry && preview != "" {
				t.Fatal("initial preview survived tool retry")
			}
			if !retry && (preview != text || len(fragments) != 2) {
				t.Fatal("tool continuation did not stream fragments")
			}
		})
	}
}

type silentFragmentProvider struct {
	mockProvider
	opened chan struct{}
}

func (p *silentFragmentProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	close(p.opened)
	return make(chan llm.StreamChunk), nil
}

func TestCompleteNarrative_FallbackDiscardsFailedPrimaryCalls(t *testing.T) {
	for _, terminal := range []string{"error", "refusal"} {
		t.Run(terminal, func(t *testing.T) {
			var ending string
			if terminal == "error" {
				ending = "event: error\ndata: {\"error\":{\"type\":\"overloaded_error\"}}\n\n"
			} else {
				ending = "event: message_delta\ndata: {\"delta\":{\"stop_reason\":\"refusal\"}}\n\nevent: message_stop\ndata: {}\n\n"
			}
			wire := strings.Join([]string{
				`event: content_block_start`, `data: {"index":0,"content_block":{"type":"text","text":""}}`, ``,
				`event: content_block_delta`, `data: {"index":0,"delta":{"type":"text_delta","text":"Abandoned text"}}`, ``,
				`event: content_block_stop`, `data: {"index":0}`, ``,
				`event: content_block_start`, `data: {"index":1,"content_block":{"type":"tool_use","id":"old","name":"update_player_hp","input":{}}}`, ``,
				`event: content_block_delta`, `data: {"index":1,"delta":{"type":"input_json_delta","partial_json":"{\"hp\":1}"}}`, ``,
				`event: content_block_stop`, `data: {"index":1}`, ``,
			}, "\n") + "\n" + ending
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/event-stream")
				_, _ = w.Write([]byte(wire))
			}))
			defer server.Close()
			secondary := &fragmentProvider{mockProvider: mockProvider{t: t}, streams: [][]llm.StreamChunk{{{ContentDelta: "Recovered "}, {ContentDelta: "narration."}, {Done: true}}}}
			provider := &llm.FallbackProvider{Primary: llm.NewClaudeClient(server.URL, "fixture", "fixture"), Secondary: secondary}
			preview := ""
			var discarded []string
			result, err := completeNarrative(context.Background(), provider, nil, nil, func(text string) {
				if text == "" {
					discarded = append(discarded, preview)
					preview = ""
				} else {
					preview += text
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			if secondary.streamCalls != 1 || preview != "Recovered narration." || result.Content != preview || len(result.ToolCalls) != 0 {
				t.Fatalf("fallback result %+v preview %q calls %d", result, preview, secondary.streamCalls)
			}
			if !strings.Contains(strings.Join(discarded, "|"), "Abandoned text") {
				t.Fatal("primary preview was not cleared")
			}
		})
	}
}
