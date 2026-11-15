package engine

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/assembly"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
)

type backpressureProvider struct{ echoStatusProvider }

func (*backpressureProvider) Complete(ctx context.Context, _ []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	calls := make([]llm.ToolCall, 64)
	for i := range calls {
		calls[i] = llm.ToolCall{ID: fmt.Sprint(i), Name: "probe", Arguments: map[string]any{}}
	}
	return &llm.Response{Content: "A quiet room.", ToolCalls: calls}, nil
}

type cancellationState struct {
	stubStateManager
	saved atomic.Int32
}

func (s *cancellationState) SaveSessionLog(context.Context, domain.SessionLog) error {
	s.saved.Add(1)
	return nil
}

func TestProcessTurnStream_CancelWithFullBuffer(t *testing.T) {
	reg := tools.NewRegistry()
	var executed atomic.Int32
	if err := reg.Register(llm.Tool{Name: "probe", Parameters: map[string]any{"type": "object"}}, func(ctx context.Context, _ map[string]any) (*tools.ToolResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		executed.Add(1)
		return &tools.ToolResult{Success: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	state := &cancellationState{}
	e := &Engine{logger: slog.Default(), state: state, assembler: assembly.NewContextAssembler(reg), processor: NewTurnProcessor(&backpressureProvider{}, reg, tools.NewValidator(reg), nil)}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := e.ProcessTurnStream(ctx, uuid.New(), "look")
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for (len(ch) < cap(ch) || executed.Load() < 7) && time.Now().Before(deadline) {
		runtime.Gosched()
	}
	if len(ch) != cap(ch) || executed.Load() != 7 {
		t.Fatal("fixture did not fill stream buffer")
	}
	before := executed.Load()
	cancel()
	// Draining allows an unsafe producer to escape too, so also assert that
	// cancellation adds at most the already in-flight send (select may race a
	// ready consumer) and starts no additional tools/persistence.
	received := 0
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				if received > cap(ch)+1 {
					t.Errorf("published %d frames after a full-buffer cancellation (capacity %d)", received, cap(ch))
				}
				if executed.Load() != before {
					t.Errorf("tool work continued after cancellation: %d -> %d", before, executed.Load())
				}
				if state.saved.Load() != 0 {
					t.Error("canceled turn reached persistence")
				}
				return
			}
			received++
		case <-timer.C:
			t.Fatal("stream did not close within one second after cancellation")
		}
	}
}

func TestStreamPublication_CancellationReleasesAllEventKinds(t *testing.T) {
	for _, kind := range []string{"status", "chunk", "result", "error"} {
		t.Run(kind, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			ch := make(chan StreamEvent, 16)
			for i := 0; i < cap(ch); i++ {
				ch <- StreamEvent{Type: "status"}
			}
			done := make(chan struct{})
			go func() { defer close(done); sendStreamEvent(ctx, ch, StreamEvent{Type: kind}) }()
			cancel()
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("publisher blocked after cancellation")
			}
			if len(ch) != cap(ch) {
				t.Fatal("fixture buffer changed")
			}
		})
	}
}
