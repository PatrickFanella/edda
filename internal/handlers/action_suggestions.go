package handlers

import (
	"errors"
	"net/http"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// SuggestActions is mounted behind campaign ownership. It never reads hidden
// facts, prompts a model, executes tools or accepts client capability claims.
func (h *CharacterHandlers) SuggestActions(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, 400, "invalid campaign id")
		return
	}
	if h.Pool == nil {
		writeError(w, 503, "action assistance unavailable")
		return
	}
	campaign := domain.Campaign{ID: campaignID}
	pc := domain.PlayerCharacter{CampaignID: campaignID}
	var location pgtype.UUID
	var sceneName string
	err = h.Pool.QueryRow(r.Context(), `SELECT c.rules_version,c.status,c.updated_at,p.id,p.name,p.stats,p.hp,p.max_hp,p.level,p.experience,p.status,p.updated_at,p.current_location_id,COALESCE(l.name,'')
 FROM campaigns c JOIN player_characters p ON p.campaign_id=c.id LEFT JOIN locations l ON l.id=p.current_location_id AND l.campaign_id=c.id AND (l.player_known OR l.player_visited) WHERE c.id=$1 ORDER BY p.created_at DESC,p.id LIMIT 1`, dbutil.ToPgtype(campaignID)).Scan(&campaign.RulesVersion, &campaign.Status, &campaign.UpdatedAt, &pc.ID, &pc.Name, &pc.Stats, &pc.HP, &pc.MaxHP, &pc.Level, &pc.Experience, &pc.Status, &pc.UpdatedAt, &location, &sceneName)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, 200, game.ActionSuggestions{Version: "suggestions-v1:empty", Candidates: []game.ActionSuggestion{}})
		return
	}
	if err != nil {
		writeError(w, 500, "action assistance unavailable")
		return
	}
	rows, err := h.Pool.Query(r.Context(), `SELECT id,fact FROM world_facts WHERE campaign_id=$1 AND player_known=TRUE AND superseded_by IS NULL AND category='clue' ORDER BY created_at DESC,id LIMIT 2`, dbutil.ToPgtype(campaignID))
	if err != nil {
		writeError(w, 500, "action assistance unavailable")
		return
	}
	defer rows.Close()
	clues := []game.KnownClue{}
	for rows.Next() {
		var clue game.KnownClue
		if err := rows.Scan(&clue.ID, &clue.Text); err != nil {
			writeError(w, 500, "action assistance unavailable")
			return
		}
		clues = append(clues, clue)
	}
	if rows.Err() != nil {
		writeError(w, 500, "action assistance unavailable")
		return
	}
	var scene *game.KnownScene
	if location.Valid {
		id := dbutil.FromPgtype(location)
		pc.CurrentLocationID = &id
		if sceneName != "" {
			scene = &game.KnownScene{ID: id, Name: sceneName}
		}
	}
	result, err := game.SuggestCharacterActions(r.Context(), campaign, pc, clues, scene)
	if err != nil {
		writeError(w, 500, "action assistance unavailable")
		return
	}
	writeJSON(w, 200, result)
}
