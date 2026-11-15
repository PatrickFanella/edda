package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"io"
	"net/http"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"github.com/google/uuid"
)

// ConfirmAction accepts a typed intent, never free-form mechanical parameters.
func (h *ActionHandlers) ConfirmAction(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, 400, "invalid campaign id")
		return
	}
	var request engine.PreparedAction
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, 400, "invalid prepared action")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		writeError(w, 400, "invalid prepared action")
		return
	}
	if request.ActionID == uuid.Nil || request.ExpectedRevision == "" {
		writeError(w, 400, "action_id and expected_revision are required")
		return
	}
	confirmer, ok := h.Engine.(interface {
		ConfirmAction(context.Context, uuid.UUID, engine.PreparedAction) (*engine.TurnResult, error)
	})
	if !ok {
		writeError(w, 503, "action confirmation unavailable")
		return
	}
	ctx, cancel := h.turnContext(r.Context())
	defer cancel()
	result, err := confirmer.ConfirmAction(ctx, campaignID, request)
	if err != nil {
		if errors.Is(err, engine.ErrActionRefresh) || errors.Is(err, engine.ErrActionIDConflict) || errors.Is(err, engine.ErrTurnInProgress) || errors.Is(err, engine.ErrTurnRecoveryRequired) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		h.Logger.Errorf("confirm action: %v", err)
		writeError(w, 500, "failed to confirm action; inspect its receipt before retrying")
		return
	}
	writeJSON(w, 200, engineTurnResultToAPI(result))
}

// confirmActionMessage shares the confirmation engine with the HTTP adapter.
func (h *ActionHandlers) confirmActionMessage(ctx context.Context, conn *websocket.Conn, campaignID uuid.UUID, payload json.RawMessage) {
	var request engine.PreparedAction
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		sendErrorEnvelope(ctx, conn, "invalid prepared action")
		return
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		sendErrorEnvelope(ctx, conn, "invalid prepared action")
		return
	}
	confirmer, ok := h.Engine.(interface {
		ConfirmAction(context.Context, uuid.UUID, engine.PreparedAction) (*engine.TurnResult, error)
	})
	if !ok {
		sendErrorEnvelope(ctx, conn, "action confirmation unavailable")
		return
	}
	turnCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	result, err := confirmer.ConfirmAction(turnCtx, campaignID, request)
	if err != nil {
		if errors.Is(err, engine.ErrActionRefresh) || errors.Is(err, engine.ErrActionIDConflict) || errors.Is(err, engine.ErrTurnInProgress) || errors.Is(err, engine.ErrTurnRecoveryRequired) {
			sendErrorEnvelope(ctx, conn, err.Error())
		} else {
			sendErrorEnvelope(ctx, conn, "failed to confirm action; inspect its receipt before retrying")
		}
		return
	}
	data, err := json.Marshal(engineTurnResultToAPI(result))
	if err != nil {
		sendErrorEnvelope(ctx, conn, "failed to read confirmed action")
		return
	}
	_ = wsjson.Write(turnCtx, conn, api.WebSocketMessageEnvelope{Type: "result", Payload: data, Timestamp: time.Now()})
}
