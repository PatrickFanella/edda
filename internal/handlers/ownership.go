package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

// RequireOwnership protects every campaign HTTP route, including WebSocket upgrades.
// Mount it after authentication on the campaign route group; engine calls from
// the local TUI do not pass through this HTTP authorization boundary.
func (h *CampaignHandlers) RequireOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := auth.UserFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		campaignID, err := campaignIDFromURL(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid campaign id")
			return
		}
		campaign, err := h.Queries.GetCampaignByID(r.Context(), db.ToPgUUID(campaignID))
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && (!campaign.CreatedBy.Valid || db.FromPgUUID(campaign.CreatedBy) != userID)) {
			writeError(w, http.StatusNotFound, "campaign not found")
			return
		}
		if err != nil {
			h.Logger.Errorf("authorize campaign %s: %v", campaignID, err)
			writeError(w, http.StatusInternalServerError, "failed to authorize campaign")
			return
		}
		// Chapter generation only reads committed player-visible history and writes
		// its own revisions. It is safe for either rules version. Other write
		// routes retain the legacy gameplay compatibility guard.
		chapterRequest := r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/journal/chapters")
		if !chapterRequest && (r.Method != http.MethodGet && r.Method != http.MethodHead || strings.EqualFold(r.Header.Get("Upgrade"), "websocket")) {
			if err := domain.RequireLegacyRules(domain.RulesVersion(campaign.RulesVersion)); err != nil {
				writeError(w, http.StatusConflict, err.Error())
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
