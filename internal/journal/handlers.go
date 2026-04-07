package journal

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handlers provides HTTP handlers for session summaries and journal entries.
type Handlers struct {
	store *Store
}

// NewHandlers creates Handlers backed by the given store.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func campaignIDFromURL(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

type summaryJSON struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	FromTurn   int    `json:"from_turn"`
	ToTurn     int    `json:"to_turn"`
	Summary    string `json:"summary"`
	CreatedAt  string `json:"created_at"`
}

func summaryToJSON(s Summary) summaryJSON {
	return summaryJSON{
		ID:         s.ID.String(),
		CampaignID: s.CampaignID.String(),
		FromTurn:   s.FromTurn,
		ToTurn:     s.ToTurn,
		Summary:    s.Summary,
		CreatedAt:  s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type entryJSON struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func entryToJSON(e Entry) entryJSON {
	return entryJSON{
		ID:         e.ID.String(),
		CampaignID: e.CampaignID.String(),
		Title:      e.Title,
		Content:    e.Content,
		CreatedAt:  e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  e.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListSummaries handles GET /api/v1/campaigns/{id}/journal/summaries.
func (h *Handlers) ListSummaries(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	summaries, err := h.store.ListSummaries(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list summaries")
		return
	}

	result := make([]summaryJSON, 0, len(summaries))
	for _, s := range summaries {
		result = append(result, summaryToJSON(s))
	}
	writeJSON(w, http.StatusOK, result)
}

// ListEntries handles GET /api/v1/campaigns/{id}/journal/entries.
func (h *Handlers) ListEntries(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	entries, err := h.store.ListEntries(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list journal entries")
		return
	}

	result := make([]entryJSON, 0, len(entries))
	for _, e := range entries {
		result = append(result, entryToJSON(e))
	}
	writeJSON(w, http.StatusOK, result)
}

type createEntryRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateEntry handles POST /api/v1/campaigns/{id}/journal/entries.
func (h *Handlers) CreateEntry(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	var req createEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	entry, err := h.store.CreateEntry(r.Context(), campaignID, req.Title, req.Content)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create journal entry")
		return
	}

	writeJSON(w, http.StatusCreated, entryToJSON(entry))
}

// DeleteEntry handles DELETE /api/v1/campaigns/{id}/journal/entries/{eid}.
func (h *Handlers) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	entryID, err := uuid.Parse(chi.URLParam(r, "eid"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid entry id")
		return
	}

	if err := h.store.DeleteEntry(r.Context(), entryID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete journal entry")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
