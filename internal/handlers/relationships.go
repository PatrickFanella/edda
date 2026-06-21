package handlers

import (
	"net/http"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// ListAwareRelationships returns player-aware relationships.
func (h *WorldHandlers) ListAwareRelationships(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	rels, err := h.Queries.ListPlayerAwareRelationships(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list relationships")
		return
	}
	resp := make([]api.RelationshipResponse, 0, len(rels))
	for _, rel := range rels {
		resp = append(resp, relationshipToResponse(rel))
	}
	writeJSON(w, http.StatusOK, resp)
}
