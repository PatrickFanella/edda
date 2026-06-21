package handlers

import (
	"net/http"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// ListKnownFacts returns player-known, non-superseded facts.
func (h *WorldHandlers) ListKnownFacts(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	facts, err := h.Queries.ListPlayerKnownFacts(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list known facts")
		return
	}
	resp := make([]api.FactResponse, 0, len(facts))
	for _, f := range facts {
		resp = append(resp, factToResponse(f))
	}
	writeJSON(w, http.StatusOK, resp)
}
