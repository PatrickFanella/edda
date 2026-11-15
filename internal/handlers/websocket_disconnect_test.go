package handlers

import (
	"context"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type disconnectEngine struct {
	wsStubEngine
	started chan struct{}
	release chan struct{}
	stopped chan struct{}
}

func (s *disconnectEngine) ProcessTurnStream(ctx context.Context, _ uuid.UUID, _ string) (<-chan engine.StreamEvent, error) {
	ch := make(chan engine.StreamEvent, 16)
	go func() {
		defer close(s.stopped)
		defer close(ch)
		close(s.started)
		select {
		case <-s.release:
		case <-ctx.Done():
			return
		}
		for {
			select {
			case ch <- engine.StreamEvent{Type: "chunk", Text: strings.Repeat("x", 16384)}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func TestHandleWebSocket_DisconnectCancelsProducer(t *testing.T) {
	for _, timeout := range []time.Duration{0, time.Minute} {
		t.Run(timeout.String(), func(t *testing.T) {
			eng := &disconnectEngine{started: make(chan struct{}), release: make(chan struct{}), stopped: make(chan struct{})}
			h := &ActionHandlers{Engine: eng, Logger: log.Default(), TurnTimeout: timeout}
			conn, _ := dialWS(t, h, uuid.New().String())
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			sendAction(t, ctx, conn, "look")
			select {
			case <-eng.started:
			case <-ctx.Done():
				t.Fatal("turn did not start")
			}
			if err := conn.CloseNow(); err != nil {
				t.Fatal(err)
			}
			close(eng.release)
			select {
			case <-eng.stopped:
			case <-ctx.Done():
				t.Fatal("producer survived failed websocket write")
			}
		})
	}
}

// silentEngine separates cancellation delivery from a dependency actually
// returning, so the test cannot mistake a stopped handler for a stopped worker.
type silentEngine struct {
	wsStubEngine
	started        chan context.Context
	canceled       chan struct{}
	providerExited chan struct{}
	workerExited   chan struct{}
	release        chan struct{}
}

func (s *silentEngine) ProcessTurnStream(ctx context.Context, _ uuid.UUID, _ string) (<-chan engine.StreamEvent, error) {
	ch := make(chan engine.StreamEvent)
	go func() {
		defer close(s.workerExited)
		defer close(ch)
		s.started <- ctx
		<-ctx.Done()
		close(s.canceled)
		if s.release != nil {
			<-s.release // Deliberately uncooperative provider; cancellation is not exit.
		}
		close(s.providerExited)
	}()
	return ch, nil
}

func TestHandleWebSocket_SilentTurnLifetime(t *testing.T) {
	for _, cause := range []string{"disconnect", "timeout", "uncooperative"} {
		t.Run(cause, func(t *testing.T) {
			s := &silentEngine{started: make(chan context.Context, 1), canceled: make(chan struct{}), providerExited: make(chan struct{}), workerExited: make(chan struct{})}
			timeout := time.Minute
			if cause == "timeout" {
				timeout = 40 * time.Millisecond
			}
			if cause == "uncooperative" {
				s.release = make(chan struct{})
			}
			h := &ActionHandlers{Engine: s, Logger: log.Default(), TurnTimeout: timeout}
			conn, srv := dialWS(t, h, uuid.NewString())
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			sendAction(t, ctx, conn, "listen")
			var turnCtx context.Context
			select {
			case turnCtx = <-s.started:
			case <-ctx.Done():
				t.Fatal("silent dependency did not start")
			}
			if cause != "timeout" {
				if err := conn.CloseNow(); err != nil {
					t.Fatal(err)
				}
			}
			select {
			case <-s.canceled:
			case <-ctx.Done():
				t.Fatal("silent dependency did not receive cancellation")
			}
			if turnCtx.Err() == nil {
				t.Fatal("turn context is still live")
			}
			if s.release != nil {
				select {
				case <-s.providerExited:
					t.Fatal("fixture exited without release")
				default:
				}
				// The HTTP worker must be free even if the provider ignores cancellation.
				srv.Close()
				close(s.release)
			}
			select {
			case <-s.providerExited:
			case <-ctx.Done():
				t.Fatal("cooperative provider did not exit")
			}
			select {
			case <-s.workerExited:
			case <-ctx.Done():
				t.Fatal("turn worker did not exit")
			}
		})
	}
}

func TestHandleWebSocket_RejectsActionDuringSilentTurn(t *testing.T) {
	s := &silentEngine{started: make(chan context.Context, 1), canceled: make(chan struct{}), providerExited: make(chan struct{}), workerExited: make(chan struct{})}
	conn, _ := dialWS(t, &ActionHandlers{Engine: s, Logger: log.Default()}, uuid.NewString())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	sendAction(t, ctx, conn, "listen")
	select {
	case <-s.started:
	case <-ctx.Done():
		t.Fatal("turn did not start")
	}
	sendAction(t, ctx, conn, "duplicate")
	env := readEnvelope(t, ctx, conn)
	if env.Type != "error" || !strings.Contains(string(env.Payload), "already in progress") {
		t.Fatalf("busy response = %+v", env)
	}
	_ = conn.CloseNow()
	select {
	case <-s.workerExited:
	case <-ctx.Done():
		t.Fatal("worker survived disconnect")
	}
}
