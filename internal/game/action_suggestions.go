package game

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"github.com/google/uuid"
)

// KnownClue is an input selected only from current player-known world facts.
type KnownClue struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

// KnownScene contains only the current location label already known to the player.
type KnownScene struct {
	ID   uuid.UUID
	Name string
}

// ActionSuggestion is optional draft assistance, never an execution promise.
type ActionSuggestion struct {
	ID        string        `json:"id"`
	Text      string        `json:"text"`
	Reason    string        `json:"reason"`
	Intent    *ActionIntent `json:"intent,omitempty"`
	SourceIDs []string      `json:"source_ids"`
}

// ActionSuggestions is a small versioned snapshot, matched locally while typing.
type ActionSuggestions struct {
	Version    string             `json:"version"`
	Candidates []ActionSuggestion `json:"candidates"`
}

// SuggestCharacterActions uses explicit known clues and read-only preparation.
// It has no provider, hidden-world lookup, dice or mutation interface. Catalogue
// entries without implemented preparation cannot become actionable suggestions.
func SuggestCharacterActions(ctx context.Context, campaign domain.Campaign, pc domain.PlayerCharacter, clues []KnownClue, scene *KnownScene) (*ActionSuggestions, error) {
	result := &ActionSuggestions{Candidates: []ActionSuggestion{}}
	if campaign.Status == domain.CampaignStatusActive && pc.Status == "active" && pc.HP > 0 {
		for _, clue := range clues {
			if len(result.Candidates) >= 2 {
				break
			}
			text := strings.TrimSpace(clue.Text)
			if text == "" {
				continue
			}
			// Bound text by runes so multilingual clues remain valid UTF-8.
			if runes := []rune(text); len(runes) > 160 {
				text = string(runes[:160]) + "…"
			}
			result.Candidates = append(result.Candidates, ActionSuggestion{ID: "clue:" + clue.ID.String(), Text: "Investigate the known clue: " + text, Reason: "You have already learned this clue. Its outcome is still uncertain.", SourceIDs: []string{"known-fact:" + clue.ID.String()}})
		}
		for _, skill := range []string{"perception", "investigation", "athletics"} {
			if len(result.Candidates) >= 4 {
				break
			}
			preview, err := PrepareCharacterAction(ctx, campaign, pc, ActionIntent{ActorID: pc.ID, Kind: "ability-check", Skill: skill})
			if err != nil {
				return nil, err
			}
			if !preview.Available {
				continue
			}
			text := map[string]string{"perception": "Look around carefully using Perception", "investigation": "Examine the scene using Investigation", "athletics": "Assess a physical approach using Athletics"}[skill]
			if scene != nil && skill == "perception" {
				text = "Look around " + scene.Name + " using Perception"
				preview.SourceIDs = append(preview.SourceIDs, "known-location:"+scene.ID.String())
			}
			intent := preview.Intent
			result.Candidates = append(result.Candidates, ActionSuggestion{ID: "skill:" + skill, Text: text, Reason: "Your current sheet supports this skill check. Circumstances and outcomes require adjudication.", Intent: &intent, SourceIDs: preview.SourceIDs})
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(struct {
		Campaign   domain.Campaign
		Character  domain.PlayerCharacter
		Candidates []ActionSuggestion
		Scene      *KnownScene
	}{campaign, pc, result.Candidates, scene})
	if err != nil {
		return nil, err
	}
	result.Version = fmt.Sprintf("suggestions-v1:%x", sha256.Sum256(encoded))
	return result, nil
}
