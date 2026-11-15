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
