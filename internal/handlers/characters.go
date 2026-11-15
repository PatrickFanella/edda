package handlers

import (
	"fmt"
	"net/http"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// GetCharacter returns the player character for a campaign.
func (h *CharacterHandlers) GetCharacter(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid campaign id: %v", err))
		return
	}

	pcs, err := h.Queries.GetPlayerCharacterByCampaign(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil {
		h.Logger.Errorf("get character for campaign %s: %v", campaignID, err)
		writeError(w, http.StatusInternalServerError, "failed to get character")
		return
	}
	if len(pcs) == 0 {
		writeError(w, http.StatusNotFound, "no character found for campaign")
		return
	}

	campaign, err := h.Queries.GetCampaignByID(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get character rules version")
		return
	}
	response := playerCharacterToResponse(pcs[0])
	response.RulesVersion = campaign.RulesVersion
	if campaign.RulesVersion == string(domain.RulesVersionSRD521) {
		pc := pcs[0]
		response.Sheet, err = srd521.Project(pc.Stats, int(pc.Level), int(pc.Hp), int(pc.MaxHp), int(pc.Experience))
		if err != nil {
			response.SheetUnavailable = err.Error()
		}
	}
	writeJSON(w, http.StatusOK, response)
}

// GetCharacterInventory returns the inventory items for the player character.
func (h *CharacterHandlers) GetCharacterInventory(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid campaign id: %v", err))
		return
	}

	pcs, err := h.Queries.GetPlayerCharacterByCampaign(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil || len(pcs) == 0 {
		writeError(w, http.StatusNotFound, "no character found for campaign")
		return
	}

	items, err := h.Queries.ListItemsByPlayer(r.Context(), statedb.ListItemsByPlayerParams{
		CampaignID:        dbutil.ToPgtype(campaignID),
		PlayerCharacterID: pcs[0].ID,
	})
	if err != nil {
		h.Logger.Errorf("list inventory for campaign %s: %v", campaignID, err)
		writeError(w, http.StatusInternalServerError, "failed to list inventory")
		return
	}

	resp := make([]api.ItemResponse, 0, len(items))
	for _, item := range items {
		resp = append(resp, itemToResponse(item))
	}
	writeJSON(w, http.StatusOK, resp)
}

// GetCharacterAbilities returns the abilities for the player character.
func (h *CharacterHandlers) GetCharacterAbilities(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid campaign id: %v", err))
		return
	}

	pcs, err := h.Queries.GetPlayerCharacterByCampaign(r.Context(), dbutil.ToPgtype(campaignID))
	if err != nil || len(pcs) == 0 {
		writeError(w, http.StatusNotFound, "no character found for campaign")
		return
	}

	writeJSON(w, http.StatusOK, parseAbilities(pcs[0].Abilities))
}
