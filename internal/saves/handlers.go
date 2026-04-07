package saves

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Handlers provides HTTP handlers for save points, campaign time, and start-over.
type Handlers struct {
	store *Store
}

// NewHandlers creates Handlers backed by the given store.
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// campaignIDFromURL extracts and parses the campaign ID from the {id} URL parameter.
func campaignIDFromURL(r *http.Request) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, "id"))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

type manualSaveRequest struct {
	Name string `json:"name"`
}

type savePointJSON struct {
	ID         string `json:"id"`
	CampaignID string `json:"campaign_id"`
	Name       string `json:"name"`
	TurnNumber int    `json:"turn_number"`
	IsAuto     bool   `json:"is_auto"`
	CreatedAt  string `json:"created_at"`
}

func savePointToJSON(sp SavePoint) savePointJSON {
	return savePointJSON{
		ID:         sp.ID.String(),
		CampaignID: sp.CampaignID.String(),
		Name:       sp.Name,
		TurnNumber: sp.TurnNumber,
		IsAuto:     sp.IsAuto,
		CreatedAt:  sp.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ManualSave handles POST /api/v1/campaigns/{id}/saves.
func (h *Handlers) ManualSave(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	var req manualSaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		req.Name = "Manual save"
	}

	turnNumber, err := h.store.GetLatestTurnNumber(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get turn number")
		return
	}

	sp, err := h.store.CreateSavePoint(r.Context(), campaignID, req.Name, turnNumber, false)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create save point")
		return
	}

	writeJSON(w, http.StatusCreated, savePointToJSON(sp))
}

// ListSaves handles GET /api/v1/campaigns/{id}/saves.
func (h *Handlers) ListSaves(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	saves, err := h.store.ListSavePoints(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list save points")
		return
	}

	result := make([]savePointJSON, 0, len(saves))
	for _, sp := range saves {
		result = append(result, savePointToJSON(sp))
	}
	writeJSON(w, http.StatusOK, result)
}

// StartOver handles POST /api/v1/campaigns/{id}/start-over.
func (h *Handlers) StartOver(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	if err := h.store.StartOver(r.Context(), campaignID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset campaign")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type campaignTimeJSON struct {
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

// GetTime handles GET /api/v1/campaigns/{id}/time.
func (h *Handlers) GetTime(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}

	ct, err := h.store.GetCampaignTime(r.Context(), campaignID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "failed to get campaign time")
		return
	}

	// If no row exists, return defaults.
	if errors.Is(err, pgx.ErrNoRows) {
		ct = CampaignTime{Day: 1, Hour: 8, Minute: 0}
	}

	writeJSON(w, http.StatusOK, campaignTimeJSON{
		Day:    ct.Day,
		Hour:   ct.Hour,
		Minute: ct.Minute,
	})
}
