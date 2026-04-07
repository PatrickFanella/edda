package export

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handlers provides HTTP handlers for campaign export.
type Handlers struct {
	store *Store
}

// NewHandlers creates Handlers backed by the given store.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func campaignIDFromURL(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// FullExport is the JSON structure for a complete campaign export.
type FullExport struct {
	Campaign   CampaignMeta      `json:"campaign"`
	Character  *ExportCharacter  `json:"character,omitempty"`
	NPCs       []ExportNPC       `json:"npcs"`
	Locations  []ExportLocation  `json:"locations"`
	Quests     []ExportQuest     `json:"quests"`
	SessionLog []ExportSessionLog `json:"session_logs"`
	WorldFacts []ExportWorldFact `json:"world_facts"`
	Inventory  []ExportItem      `json:"inventory"`
}

// ExportJSON handles GET /api/v1/campaigns/{id}/export/json.
func (h *Handlers) ExportJSON(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	ctx := r.Context()

	meta, err := h.store.GetCampaignMeta(ctx, campaignID)
	if err != nil {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}

	export := FullExport{Campaign: meta}

	pc, err := h.store.GetPlayerCharacter(ctx, campaignID)
	if err == nil {
		export.Character = &pc
	}

	export.NPCs, _ = h.store.ListAllNPCs(ctx, campaignID)
	if export.NPCs == nil {
		export.NPCs = []ExportNPC{}
	}
	export.Locations, _ = h.store.ListAllLocations(ctx, campaignID)
	if export.Locations == nil {
		export.Locations = []ExportLocation{}
	}
	export.Quests, _ = h.store.ListAllQuests(ctx, campaignID)
	if export.Quests == nil {
		export.Quests = []ExportQuest{}
	}
	export.SessionLog, _ = h.store.ListAllSessionLogs(ctx, campaignID)
	if export.SessionLog == nil {
		export.SessionLog = []ExportSessionLog{}
	}
	export.WorldFacts, _ = h.store.ListAllWorldFacts(ctx, campaignID)
	if export.WorldFacts == nil {
		export.WorldFacts = []ExportWorldFact{}
	}
	export.Inventory, _ = h.store.ListPlayerItems(ctx, campaignID)
	if export.Inventory == nil {
		export.Inventory = []ExportItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="campaign-%s.json"`, campaignID))
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(export)
}

// ExportTranscript handles GET /api/v1/campaigns/{id}/export/transcript.
func (h *Handlers) ExportTranscript(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	ctx := r.Context()

	meta, err := h.store.GetCampaignMeta(ctx, campaignID)
	if err != nil {
		writeError(w, http.StatusNotFound, "campaign not found")
		return
	}

	logs, err := h.store.ListAllSessionLogs(ctx, campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load session logs")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s — Session Transcript\n\n", meta.Name)

	for _, log := range logs {
		fmt.Fprintf(&sb, "## Turn %d\n", log.TurnNumber)
		fmt.Fprintf(&sb, "**Player:** %s\n\n", log.PlayerInput)
		sb.WriteString(log.LLMResponse)
		sb.WriteString("\n\n---\n\n")
	}

	if len(logs) == 0 {
		sb.WriteString("_No session logs recorded yet._\n")
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="transcript-%s.md"`, campaignID))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}

// ExportCharacterSheet handles GET /api/v1/campaigns/{id}/export/character.
func (h *Handlers) ExportCharacterSheet(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	ctx := r.Context()

	pc, err := h.store.GetPlayerCharacter(ctx, campaignID)
	if err != nil {
		writeError(w, http.StatusNotFound, "character not found")
		return
	}

	items, _ := h.store.ListPlayerItems(ctx, campaignID)
	if items == nil {
		items = []ExportItem{}
	}

	// Parse abilities from JSON.
	type ability struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}
	var abilities []ability
	_ = json.Unmarshal(pc.Abilities, &abilities)

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", pc.Name)
	fmt.Fprintf(&sb, "**Level:** %d | **Status:** %s\n\n", pc.Level, pc.Status)

	if pc.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", pc.Description)
	}

	sb.WriteString("## Stats\n")
	fmt.Fprintf(&sb, "- HP: %d/%d\n", pc.HP, pc.MaxHP)
	fmt.Fprintf(&sb, "- XP: %d\n", pc.Experience)
	sb.WriteString("\n")

	if len(abilities) > 0 {
		sb.WriteString("## Abilities\n")
		for _, a := range abilities {
			if a.Description != "" {
				fmt.Fprintf(&sb, "- **%s**: %s\n", a.Name, a.Description)
			} else {
				fmt.Fprintf(&sb, "- %s\n", a.Name)
			}
		}
		sb.WriteString("\n")
	}

	if len(items) > 0 {
		sb.WriteString("## Inventory\n")
		sb.WriteString("| Item | Qty | Equipped |\n")
		sb.WriteString("|------|-----|----------|\n")
		for _, item := range items {
			equipped := "No"
			if item.Equipped {
				equipped = "Yes"
			}
			fmt.Fprintf(&sb, "| %s | %d | %s |\n", item.Name, item.Quantity, equipped)
		}
		sb.WriteString("\n")
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="character-%s.md"`, campaignID))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(sb.String()))
}
