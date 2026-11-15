package journal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/db"
	"github.com/google/uuid"
)

// ReminderItem quotes a current player-visible source, without interpretation.
type ReminderItem struct {
	Text     string `json:"text"`
	SourceID string `json:"source_id"`
}

// ReturnReminder is a bounded projection of committed player-known state.
// It deliberately excludes generated summaries, hidden lore and future objectives.
type ReturnReminder struct {
	Scene      *ReminderItem  `json:"scene,omitempty"`
	Objective  *ReminderItem  `json:"objective,omitempty"`
	KnownFacts []ReminderItem `json:"known_facts"`
}

// One statement gives all fields the same database snapshot. Quest titles are
// already player-facing in the quest log; descriptions/future objectives are not
// needed here. A location must be explicitly known or visited, even if current.
const returnReminderSQL = `
WITH current_character AS (
    SELECT current_location_id FROM player_characters
    WHERE campaign_id=$1 ORDER BY created_at DESC,id LIMIT 1
), current_quest AS (
    SELECT id,title FROM quests
    WHERE campaign_id=$1 AND status='active' AND btrim(title)<>''
    ORDER BY updated_at DESC,id LIMIT 1
), known_facts AS (
    SELECT id,fact FROM world_facts
    WHERE campaign_id=$1 AND player_known AND superseded_by IS NULL AND btrim(fact)<>''
    ORDER BY (category='clue') DESC,created_at DESC,id LIMIT 3
)
SELECT 'scene',l.id::text,l.name FROM locations l
JOIN current_character p ON p.current_location_id=l.id
WHERE l.campaign_id=$1 AND (l.player_known OR l.player_visited) AND btrim(l.name)<>''
UNION ALL SELECT 'objective',id::text,title FROM current_quest
UNION ALL SELECT 'fact',id::text,fact FROM known_facts
ORDER BY 1 DESC,2
`

// ReadReturnReminder never calls a provider or writes gameplay/history. Missing
// knowledge stays missing rather than being filled from a model or world state.
func (s *Store) ReadReturnReminder(ctx context.Context, campaignID uuid.UUID) (*ReturnReminder, error) {
	rows, err := s.db.Query(ctx, returnReminderSQL, db.ToPgUUID(campaignID))
	if err != nil {
		return nil, fmt.Errorf("read return reminder: %w", err)
	}
	defer rows.Close()
	result := &ReturnReminder{KnownFacts: []ReminderItem{}}
	for rows.Next() {
		var kind, id, text string
		if err := rows.Scan(&kind, &id, &text); err != nil {
			return nil, fmt.Errorf("scan return reminder: %w", err)
		}
		item := ReminderItem{Text: reminderText(text)}
		switch kind {
		case "scene":
			item.SourceID = "known-location:" + id
			result.Scene = &item
		case "objective":
			item.SourceID = "quest:" + id
			result.Objective = &item
		case "fact":
			item.SourceID = "known-fact:" + id
			result.KnownFacts = append(result.KnownFacts, item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read return reminder rows: %w", err)
	}
	return result, nil
}

func reminderText(text string) string {
	// Bound by runes to preserve multilingual text. Do not paraphrase a fact.
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > 200 {
		return string(runes[:200]) + "…"
	}
	return string(runes)
}

// GetReturnReminder handles the ownership-guarded GET journal/reminder route.
func (h *Handlers) GetReturnReminder(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	reminder, err := h.store.ReadReturnReminder(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "return reminder unavailable")
		return
	}
	writeJSON(w, http.StatusOK, reminder)
}
