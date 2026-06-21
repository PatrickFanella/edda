package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

const testJWTSecret = "test-secret"

// wsStubEngine implements engine.GameEngine with controllable ProcessTurnStream.
type wsStubEngine struct {
	streamEvents []engine.StreamEvent
	streamErr    error
}

func (s *wsStubEngine) ProcessTurn(_ context.Context, _ uuid.UUID, _ string) (*engine.TurnResult, error) {
	return nil, nil
}

func (s *wsStubEngine) GetGameState(_ context.Context, _ uuid.UUID) (*engine.GameState, error) {
	return nil, nil
}

func (s *wsStubEngine) NewCampaign(_ context.Context, _ uuid.UUID) (*domain.Campaign, error) {
	return nil, nil
}

func (s *wsStubEngine) LoadCampaign(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (s *wsStubEngine) ProcessTurnStream(_ context.Context, _ uuid.UUID, _ string) (<-chan engine.StreamEvent, error) {
	if s.streamErr != nil {
		return nil, s.streamErr
	}
	ch := make(chan engine.StreamEvent, len(s.streamEvents))
	for _, e := range s.streamEvents {
		ch <- e
	}
	close(ch)
	return ch, nil
}

func newWSRouter(h *ActionHandlers) *chi.Mux {
	r := chi.NewRouter()
	authMW := auth.NewNoOpMiddleware(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
	r.Use(authMW.Authenticate)
	r.Route("/campaigns/{id}", func(r chi.Router) {
		r.Get("/ws", h.HandleWebSocket)
	})
	return r
}

// dialWS starts a test server and dials its WebSocket endpoint.
func dialWS(t *testing.T, h *ActionHandlers, campaignID string) (*websocket.Conn, *httptest.Server) {
	t.Helper()
	return dialWSWithOptions(t, h, campaignID, nil)
}

func dialWSWithOptions(t *testing.T, h *ActionHandlers, campaignID string, opts *websocket.DialOptions) (*websocket.Conn, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(newWSRouter(h))
	t.Cleanup(srv.Close)

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/campaigns/" + campaignID + "/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	conn, _, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })

	return conn, srv
}

// sendAction writes an action message to the WebSocket.
func sendAction(t *testing.T, ctx context.Context, conn *websocket.Conn, input string) {
	t.Helper()
	payload, _ := json.Marshal(map[string]string{"input": input})
	msg := map[string]any{
		"type":    "action",
		"payload": json.RawMessage(payload),
	}
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		t.Fatalf("write action: %v", err)
	}
}

// readEnvelope reads a single WebSocketMessageEnvelope.
func readEnvelope(t *testing.T, ctx context.Context, conn *websocket.Conn) api.WebSocketMessageEnvelope {
	t.Helper()
	var env api.WebSocketMessageEnvelope
	if err := wsjson.Read(ctx, conn, &env); err != nil {
		t.Fatalf("read envelope: %v", err)
	}
	return env
}

func TestHandleWebSocket_Success(t *testing.T) {
	campaignID := uuid.New()
	playerID := uuid.New()
	locationID := uuid.New()
	combatStateID := uuid.New()
	moveApplied := engine.AppliedToolCall{Tool: "move_player", Result: mustJSON(t, map[string]any{"location_id": locationID.String(), "campaign_id": campaignID.String(), "player_character_id": playerID.String(), "name": "Old Road", "description": "A weathered road", "location_type": "wilderness", "travel_time": "1 hour", "day": 1, "hour": 9, "minute": 0, "visited_marked": true, "visited_warning": "", "time_warning": ""})}
	combatApplied := engine.AppliedToolCall{Tool: "resolve_combat", Result: mustJSON(t, map[string]any{"xp_earned": 150, "loot": []any{}, "dead_npc_ids": []any{}, "combat_state": map[string]any{"id": combatStateID.String(), "campaign_id": campaignID.String(), "round_number": 2, "status": "completed", "narrative": "", "initiative_order": []any{}, "environment": map[string]any{"description": "trail"}, "combatants": []any{}}, "outcome_type": "victory"})}
	eng := &wsStubEngine{
		streamEvents: []engine.StreamEvent{
			{Type: "status", Status: &api.StatusPayload{Stage: "thinking", Description: "Generating response..."}},
			{Type: "chunk", Text: "You see a dragon."},
			{Type: "result", Result: &engine.TurnResult{
				Narrative:    "You see a dragon.",
				StateChanges: append(engine.StateChangesFromAppliedToolCalls([]engine.AppliedToolCall{moveApplied}), engine.StateChangesFromAppliedToolCalls([]engine.AppliedToolCall{combatApplied})...),
			}},
		},
	}
	h := &ActionHandlers{Engine: eng, Logger: log.Default()}
	campaignIDStr := campaignID.String()
	conn, _ := dialWS(t, h, campaignIDStr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sendAction(t, ctx, conn, "look around")

	// Expect status envelope.
	env := readEnvelope(t, ctx, conn)
	if env.Type != "status" {
		t.Fatalf("expected type status, got %q", env.Type)
	}
	var statusPayload api.StatusPayload
	if err := json.Unmarshal(env.Payload, &statusPayload); err != nil {
		t.Fatalf("unmarshal status payload: %v", err)
	}
	if statusPayload.Stage != "thinking" {
		t.Fatalf("expected thinking stage, got %q", statusPayload.Stage)
	}

	// Expect chunk envelope.
	env = readEnvelope(t, ctx, conn)
	if env.Type != "chunk" {
		t.Fatalf("expected type chunk, got %q", env.Type)
	}
	var chunkPayload map[string]string
	if err := json.Unmarshal(env.Payload, &chunkPayload); err != nil {
		t.Fatalf("unmarshal chunk payload: %v", err)
	}
	if chunkPayload["text"] != "You see a dragon." {
		t.Errorf("expected chunk text %q, got %q", "You see a dragon.", chunkPayload["text"])
	}

	// Expect result envelope.
	env = readEnvelope(t, ctx, conn)
	if env.Type != "result" {
		t.Fatalf("expected type result, got %q", env.Type)
	}
	if !strings.Contains(string(env.Payload), `"state_changes"`) {
		t.Fatalf("expected snake_case state_changes key in payload: %s", string(env.Payload))
	}
	if strings.Contains(string(env.Payload), `"StateChanges"`) {
		t.Fatalf("unexpected PascalCase StateChanges key in payload: %s", string(env.Payload))
	}
	var result api.TurnResponse
	if err := json.Unmarshal(env.Payload, &result); err != nil {
		t.Fatalf("unmarshal result payload: %v", err)
	}
	if result.Narrative != "You see a dragon." {
		t.Errorf("expected narrative %q, got %q", "You see a dragon.", result.Narrative)
	}
	if len(result.StateChanges) != 4 || result.StateChanges[0].EntityType != "player_character" || result.StateChanges[0].ChangeType != "location_updated" || result.StateChanges[1].EntityType != "location" || result.StateChanges[1].ChangeType != "updated" || result.StateChanges[2].EntityType != "location" || result.StateChanges[2].ChangeType != "moved" || result.StateChanges[3].EntityType != "combat" || result.StateChanges[3].ChangeType != "resolved" {
		t.Fatalf("expected converted state changes, got %+v", result.StateChanges)
	}
	if _, ok := result.StateChanges[3].Details["outcome_type"]; ok {
		t.Fatalf("did not expect outcome_type to be projected, got %+v", result.StateChanges[3].Details)
	}
}

func TestHandleWebSocket_AllowsFrontendDevOrigin(t *testing.T) {
	h := &ActionHandlers{Engine: &wsStubEngine{}, Logger: log.Default()}
	conn, _ := dialWSWithOptions(t, h, uuid.New().String(), &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Origin": []string{"http://127.0.0.1:5173"},
		},
	})

	if conn == nil {
		t.Fatal("expected websocket connection for allowed frontend origin")
	}
}

func TestHandleWebSocket_InvalidCampaignID(t *testing.T) {
	h := &ActionHandlers{Engine: &wsStubEngine{}, Logger: log.Default()}
	srv := httptest.NewServer(newWSRouter(h))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/campaigns/not-a-uuid/ws"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Dial should fail because the server responds with HTTP 400 before upgrade.
	_, _, err := websocket.Dial(ctx, wsURL, nil)
	if err == nil {
		t.Fatal("expected dial to fail for invalid campaign ID")
	}
}

func TestHandleWebSocket_EmptyInput(t *testing.T) {
	eng := &wsStubEngine{}
	h := &ActionHandlers{Engine: eng, Logger: log.Default()}
	campaignID := uuid.New().String()
	conn, _ := dialWS(t, h, campaignID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sendAction(t, ctx, conn, "")

	env := readEnvelope(t, ctx, conn)
	if env.Type != "error" {
		t.Fatalf("expected type error, got %q", env.Type)
	}
	var errPayload map[string]string
	if err := json.Unmarshal(env.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload["error"] != "input is required" {
		t.Errorf("expected error %q, got %q", "input is required", errPayload["error"])
	}
}

func TestHandleWebSocket_StreamError(t *testing.T) {
	eng := &wsStubEngine{
		streamErr: errors.New("llm unavailable"),
	}
	h := &ActionHandlers{Engine: eng, Logger: log.Default()}
	campaignID := uuid.New().String()
	conn, _ := dialWS(t, h, campaignID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sendAction(t, ctx, conn, "attack")

	env := readEnvelope(t, ctx, conn)
	if env.Type != "error" {
		t.Fatalf("expected type error, got %q", env.Type)
	}
	var errPayload map[string]string
	if err := json.Unmarshal(env.Payload, &errPayload); err != nil {
		t.Fatalf("unmarshal error payload: %v", err)
	}
	if errPayload["error"] != "failed to process turn" {
		t.Errorf("expected error %q, got %q", "failed to process turn", errPayload["error"])
	}
}

func TestHandleWebSocket_GracefulClose(t *testing.T) {
	eng := &wsStubEngine{}
	h := &ActionHandlers{Engine: eng, Logger: log.Default()}
	campaignID := uuid.New().String()
	conn, _ := dialWS(t, h, campaignID)

	// Close normally — handler should not panic or log an error.
	err := conn.Close(websocket.StatusNormalClosure, "bye")
	if err != nil {
		t.Fatalf("close websocket: %v", err)
	}
}

func TestHandleWebSocket_StreamErrorEvent(t *testing.T) {
	t.Run("NilErr", func(t *testing.T) {
		eng := &wsStubEngine{
			streamEvents: []engine.StreamEvent{
				{Type: "error", Err: nil},
			},
		}
		h := &ActionHandlers{Engine: eng, Logger: log.Default()}
		campaignID := uuid.New().String()
		conn, _ := dialWS(t, h, campaignID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sendAction(t, ctx, conn, "look")

		env := readEnvelope(t, ctx, conn)
		if env.Type != "error" {
			t.Fatalf("expected type error, got %q", env.Type)
		}
		var errPayload map[string]string
		if err := json.Unmarshal(env.Payload, &errPayload); err != nil {
			t.Fatalf("unmarshal error payload: %v", err)
		}
		if errPayload["error"] != "an internal error occurred" {
			t.Errorf("expected generic error, got %q", errPayload["error"])
		}
	})

	t.Run("WithErr", func(t *testing.T) {
		eng := &wsStubEngine{
			streamEvents: []engine.StreamEvent{
				{Type: "error", Err: errors.New("db timeout")},
			},
		}
		h := &ActionHandlers{Engine: eng, Logger: log.Default()}
		campaignID := uuid.New().String()
		conn, _ := dialWS(t, h, campaignID)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		sendAction(t, ctx, conn, "look")

		env := readEnvelope(t, ctx, conn)
		if env.Type != "error" {
			t.Fatalf("expected type error, got %q", env.Type)
		}
		var errPayload map[string]string
		if err := json.Unmarshal(env.Payload, &errPayload); err != nil {
			t.Fatalf("unmarshal error payload: %v", err)
		}
		if errPayload["error"] != "an internal error occurred" {
			t.Errorf("expected generic error, got %q", errPayload["error"])
		}
		if strings.Contains(errPayload["error"], "db timeout") {
			t.Errorf("error leaked internal details: %q", errPayload["error"])
		}
	})
}

func TestHandleWebSocket_AllowsCookieBackedJWTUpgrade(t *testing.T) {
	eng := &wsStubEngine{
		streamEvents: []engine.StreamEvent{{Type: "chunk", Text: "You see a dragon."}},
	}
	h := &ActionHandlers{Engine: eng, Logger: log.Default()}
	campaignID := uuid.New().String()
	userID := uuid.New()
	token, err := auth.GenerateToken(userID, testJWTSecret, auth.DefaultTokenTTL)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	r := chi.NewRouter()
	r.Use(auth.NewJWTMiddleware(testJWTSecret).Authenticate)
	r.Route("/campaigns/{id}", func(r chi.Router) {
		r.Get("/ws", h.HandleWebSocket)
	})

	srv := httptest.NewServer(r)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/campaigns/" + campaignID + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: http.Header{
			"Origin": []string{"http://127.0.0.1:5173"},
			"Cookie": []string{auth.AuthCookieName + "=" + token},
		},
	})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer func() { _ = conn.CloseNow() }()

	sendAction(t, ctx, conn, "look around")
	env := readEnvelope(t, ctx, conn)
	if env.Type != "chunk" {
		t.Fatalf("expected type chunk, got %q", env.Type)
	}
}
