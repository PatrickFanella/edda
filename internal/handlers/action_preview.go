package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/game"
)

// PreviewAction reads a single statement snapshot after campaign authorization.
// GET is deliberate: preparation has no dice, mutation or confirmation path.
func (h *CharacterHandlers) PreviewAction(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	userID, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	values := r.URL.Query()
	for name, entries := range values {
		if len(entries) != 1 || (name != "actor_id" && name != "kind" && name != "ability" && name != "skill" && name != "entry_id" && name != "slot_level") {
			writeError(w, http.StatusBadRequest, "invalid action preview parameters")
			return
		}
	}
	slotLevel := 0
	if raw := values.Get("slot_level"); raw != "" {
		var err error
		slotLevel, err = strconv.Atoi(raw)
		if err != nil || slotLevel < 1 || slotLevel > 9 {
			writeError(w, http.StatusBadRequest, "invalid slot level")
			return
		}
	}
	actorID, err := uuid.Parse(values.Get("actor_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid actor id")
		return
	}
	if h.Pool == nil {
		writeError(w, http.StatusServiceUnavailable, "action previews unavailable")
		return
	}
	campaign, pc, err := game.ReadActionSnapshot(r.Context(), h.Pool, campaignID, actorID, userID, false)
	if errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusNotFound, "character action not found")
		return
	}
	if err != nil {
		h.Logger.Errorf("read action preview: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to prepare action preview")
		return
	}
	prepared, err := game.PrepareStoredAction(r.Context(), h.Pool, campaign, pc, game.ActionIntent{ActorID: actorID, Kind: values.Get("kind"), Ability: values.Get("ability"), Skill: values.Get("skill"), EntryID: values.Get("entry_id"), SlotLevel: slotLevel}, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare action preview")
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, prepared.Preview)
}
