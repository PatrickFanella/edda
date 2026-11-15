package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net/http"

	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// ProcessAction handles a player turn action.
func (h *ActionHandlers) ProcessAction(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid campaign id: %v", err))
		return
	}

	var req api.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Input == "" {
		writeError(w, http.StatusBadRequest, "input is required")
		return
	}

	turnCtx, cancel := h.turnContext(r.Context())
	defer cancel()

	turnCtx, err = actionContext(turnCtx, req.ActionID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid action_id")
		return
	}
	result, err := h.Engine.ProcessTurn(turnCtx, campaignID, req.Input)
	if err != nil {
		h.Logger.Errorf("process turn for campaign %s: %v", campaignID, err)
		if errors.Is(err, engine.ErrTurnRecoveryRequired) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": err.Error(), "action_id": engine.RecoveryActionID(err)})
			return
		}
		if errors.Is(err, engine.ErrTurnInProgress) || errors.Is(err, engine.ErrActionIDConflict) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		if engine.IsLLMOutputError(err) {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to process turn")
		return
	}

	writeJSON(w, http.StatusOK, engineTurnResultToAPI(result))
}

func (h *ActionHandlers) turnContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(parent)
	if h.TurnTimeout <= 0 {
		return context.WithCancel(base)
	}
	return context.WithTimeout(base, h.TurnTimeout)
}

func actionContext(ctx context.Context, raw string) (context.Context, error) {
	if raw == "" {
		return ctx, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return nil, fmt.Errorf("invalid action id")
	}
	return engine.WithActionID(ctx, id), nil
}

// GetActionReceipt reads a pending action, or a specific action_id. Ownership
// is enforced by the same campaign middleware as ProcessAction.
func (h *ActionHandlers) GetActionReceipt(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	actionID := uuid.Nil
	if raw := r.URL.Query().Get("action_id"); raw != "" {
		actionID, err = uuid.Parse(raw)
		if err != nil || actionID == uuid.Nil {
			writeError(w, http.StatusBadRequest, "invalid action_id")
			return
		}
	}
	reader, ok := h.Engine.(interface {
		GetActionReceipt(context.Context, uuid.UUID, uuid.UUID) (*engine.ActionReceipt, error)
	})
	if !ok {
		writeError(w, http.StatusNotImplemented, "action receipts unavailable")
		return
	}
	receipt, err := reader.GetActionReceipt(r.Context(), campaignID, actionID)
	if err != nil {
		h.Logger.Errorf("read action receipt for campaign %s: %v", campaignID, err)
		writeError(w, http.StatusInternalServerError, "failed to read action receipt")
		return
	}
	if receipt == nil {
		writeJSON(w, http.StatusOK, nil)
		return
	}
	projection := engineTurnResultToAPI(&engine.TurnResult{AppliedToolCalls: receipt.AppliedToolCalls, StateChanges: engine.StateChangesFromAppliedToolCalls(receipt.AppliedToolCalls)})
	response := api.ActionReceiptResponse{
		ActionID: receipt.ActionID.String(), Input: receipt.PlayerInput,
		Status: receipt.Status, TurnNumber: receipt.TurnNumber,
		NarrationAttempts: receipt.NarrationAttempts, MaxNarrationAttempts: engine.MaxNarrationAttempts,
		StateChanges: projection.StateChanges, ResolutionEvents: projection.ResolutionEvents,
	}
	if receipt.Result != nil {
		result := engineTurnResultToAPI(receipt.Result)
		response.Result = &result
	}
	writeJSON(w, http.StatusOK, response)
}
