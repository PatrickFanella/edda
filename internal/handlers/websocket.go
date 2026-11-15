package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// wsActionMessage is the client-to-server message envelope.
type wsActionMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// wsActionPayload carries the player input within an action message.
type wsActionPayload = api.ActionRequest

var websocketAcceptOptions = &websocket.AcceptOptions{
	OriginPatterns: []string{
		"localhost:*",
		"127.0.0.1:*",
		"https://edda.subcult.tv",
	},
}

// HandleWebSocket upgrades to WebSocket and streams game events to the client.
func (h *ActionHandlers) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid campaign id: %v", err))
		return
	}

	conn, err := websocket.Accept(w, r, websocketAcceptOptions)
	if err != nil {
		h.Logger.Errorf("websocket accept for campaign %s: %v", campaignID, err)
		return
	}
	defer func() { _ = conn.CloseNow() }()

	ctx, cancelConnection := context.WithCancel(r.Context())
	defer cancelConnection()
	// Exactly one reader stays active while the provider is silent. A peer close
	// cancels turn work without waiting for a narrative write to discover it.
	messages := make(chan wsActionMessage)
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		defer cancelConnection()
		for {
			var msg wsActionMessage
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				return
			}
			select {
			case messages <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()
	defer func() {
		cancelConnection()
		_ = conn.CloseNow()
		<-readerDone
	}()

	for {
		var msg wsActionMessage
		select {
		case <-ctx.Done():
			return
		case msg = <-messages:
		}

		if msg.Type != "action" {
			sendErrorEnvelope(ctx, conn, fmt.Sprintf("unknown message type: %q", msg.Type))
			continue
		}

		var payload wsActionPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			sendErrorEnvelope(ctx, conn, fmt.Sprintf("invalid payload: %v", err))
			continue
		}

		if payload.Input == "" {
			sendErrorEnvelope(ctx, conn, "input is required")
			continue
		}

		turnCtx, cancelTurn := h.turnContext(ctx)
		stopCancellation := context.AfterFunc(ctx, cancelTurn)
		turnCtx, err = actionContext(turnCtx, payload.ActionID)
		if err != nil {
			stopCancellation()
			cancelTurn()
			sendErrorEnvelope(ctx, conn, "invalid action_id")
			continue
		}
		ch, err := h.Engine.ProcessTurnStream(turnCtx, campaignID, payload.Input)
		if err != nil {
			stopCancellation()
			cancelTurn()
			h.Logger.Errorf("process turn stream for campaign %s: %v", campaignID, err)
			sendErrorEnvelope(ctx, conn, "failed to process turn")
			continue
		}

	stream:
		for {
			var event engine.StreamEvent
			var ok bool
			select {
			case <-ctx.Done():
				stopCancellation()
				cancelTurn()
				return
			case <-turnCtx.Done():
				sendErrorEnvelope(ctx, conn, "turn canceled or timed out; reconnect to inspect its action receipt")
				break stream
			case <-messages:
				sendErrorEnvelope(ctx, conn, "a turn is already in progress; inspect its action receipt before retrying")
				continue
			case event, ok = <-ch:
				if !ok {
					break stream
				}
			}
			var envelope api.WebSocketMessageEnvelope
			envelope.Timestamp = time.Now()

			switch event.Type {
			case "chunk":
				envelope.Type = "chunk"
				envelope.Payload, _ = json.Marshal(map[string]string{"text": event.Text})
			case "result":
				envelope.Type = "result"
				if event.Result == nil {
					envelope.Type = "error"
					envelope.Payload, _ = json.Marshal(map[string]string{"error": "an internal error occurred"})
					break
				}
				resultPayload, err := json.Marshal(engineTurnResultToAPI(event.Result))
				if err != nil {
					h.Logger.Errorf("marshal turn result for campaign %s: %v", campaignID, err)
					envelope.Type = "error"
					envelope.Payload, _ = json.Marshal(map[string]string{"error": "an internal error occurred"})
					break
				}
				envelope.Payload = resultPayload
			case "status":
				envelope.Type = "status"
				if event.Status != nil {
					envelope.Payload, _ = json.Marshal(event.Status)
				} else {
					continue
				}
			case "error":
				envelope.Type = "error"
				errMsg := "an internal error occurred"
				if event.Err != nil {
					h.Logger.Errorf("stream event error for campaign %s: %v", campaignID, event.Err)
				}
				if errors.Is(event.Err, engine.ErrActionIDConflict) {
					errMsg = event.Err.Error()
				}
				if errors.Is(event.Err, engine.ErrTurnRecoveryRequired) {
					envelope.Payload, _ = json.Marshal(map[string]string{"error": event.Err.Error(), "action_id": engine.RecoveryActionID(event.Err)})
				} else {
					envelope.Payload, _ = json.Marshal(map[string]string{"error": errMsg})
				}
			default:
				continue
			}

			if err := wsjson.Write(ctx, conn, envelope); err != nil {
				stopCancellation()
				cancelTurn()
				h.Logger.Errorf("websocket write for campaign %s: %v", campaignID, err)
				return
			}
		}
		stopCancellation()
		cancelTurn()
	}
}

// sendErrorEnvelope writes an error envelope to the WebSocket connection.
func sendErrorEnvelope(ctx context.Context, conn *websocket.Conn, msg string) {
	payload, _ := json.Marshal(map[string]string{"error": msg})
	envelope := api.WebSocketMessageEnvelope{
		Type:      "error",
		Payload:   payload,
		Timestamp: time.Now(),
	}
	//nolint:errcheck // Best-effort; connection may already be closing.
	wsjson.Write(ctx, conn, envelope)
}
