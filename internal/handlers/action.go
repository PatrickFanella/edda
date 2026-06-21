package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	result, err := h.Engine.ProcessTurn(turnCtx, campaignID, req.Input)
	if err != nil {
		h.Logger.Errorf("process turn for campaign %s: %v", campaignID, err)
		writeError(w, http.StatusInternalServerError, "failed to process turn")
		return
	}

	writeJSON(w, http.StatusOK, engineTurnResultToAPI(result))
}

func (h *ActionHandlers) turnContext(parent context.Context) (context.Context, context.CancelFunc) {
	base := context.WithoutCancel(parent)
	if h.TurnTimeout <= 0 {
		return base, func() {}
	}
	return context.WithTimeout(base, h.TurnTimeout)
}
