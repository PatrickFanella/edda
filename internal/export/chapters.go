package export

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

// ChapterExport contains only selected saved prose and its player-visible source.
// It deliberately does not embed FullExport or current campaign/world metadata.
type ChapterExport struct {
	CampaignID  string                `json:"campaign_id"`
	Chapters    []api.ChapterResponse `json:"chapters"`
	Attribution string                `json:"attribution"`
	Changes     string                `json:"changes"`
}

const chapterExportChanges = "Edda chapters are saved generated prose based on the captured player-visible transcript. Rules references are adapted for Edda; this export does not regenerate prose or reinterpret recorded outcomes."

// ExportChapters handles GET /api/v1/campaigns/{id}/export/chapters behind
// the same authentication and campaign ownership middleware as other exports.
// Repeated revision_id parameters select exact completed revisions, never latest
// at download time. A selection that is no longer available fails as a whole.
func (h *Handlers) ExportChapters(w http.ResponseWriter, r *http.Request) {
	campaignID, err := campaignIDFromURL(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid campaign id")
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "json" {
		writeError(w, http.StatusBadRequest, "format must be markdown or json")
		return
	}
	ids := r.URL.Query()["revision_id"]
	if len(ids) == 0 || len(ids) > 100 {
		writeError(w, http.StatusBadRequest, "select between 1 and 100 saved revisions")
		return
	}
	selected := make(map[string]bool, len(ids))
	for _, value := range ids {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || selected[id.String()] {
			writeError(w, http.StatusBadRequest, "revision_id must be a unique nonzero UUID")
			return
		}
		selected[id.String()] = true
	}
	// Reuse the reader's saved source/revision projection, with no generation
	// provider, current transcript, hidden lore, or raw tool data in this path.
	chapters, err := journal.NewStore(h.store.db).ListChapters(r.Context(), campaignID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load saved chapters")
		return
	}
	chapters, ok := selectedChapters(chapters, selected)
	if !ok {
		writeError(w, http.StatusNotFound, "selected saved revision not available")
		return
	}
	payload := ChapterExport{CampaignID: campaignID.String(), Chapters: chapters, Attribution: srd521.Attribution, Changes: chapterExportChanges}
	w.Header().Set("Cache-Control", "private, no-store")
	if format == "json" {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="chapters-%s.json"`, campaignID))
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(payload)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="chapters-%s.md"`, campaignID))
	_, _ = w.Write([]byte(chapterMarkdown(payload)))
}

func selectedChapters(chapters []api.ChapterResponse, selected map[string]bool) ([]api.ChapterResponse, bool) {
	result := []api.ChapterResponse{}
	found := 0
	for _, chapter := range chapters {
		revisions := []api.ChapterRevisionResponse{}
		for _, revision := range chapter.Revisions {
			if selected[revision.ID] && revision.Status == "completed" {
				revisions = append(revisions, revision)
				found++
			}
		}
		if len(revisions) > 0 {
			chapter.Revisions = revisions
			result = append(result, chapter)
		}
	}
	// Match the reader, including retained chapters whose turn numbers restarted.
	sort.Slice(result, func(i, j int) bool {
		a, b := result[i], result[j]
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.Before(b.CreatedAt)
		}
		if a.FromTurn != b.FromTurn {
			return a.FromTurn < b.FromTurn
		}
		if a.ToTurn != b.ToTurn {
			return a.ToTurn < b.ToTurn
		}
		return a.ID < b.ID
	})
	return result, found == len(selected)
}

func chapterMarkdown(export ChapterExport) string {
	var out strings.Builder
	fmt.Fprintf(&out, "# Saved campaign chapters\n\nCampaign: %s\n\n", export.CampaignID)
	for i, chapter := range export.Chapters {
		fmt.Fprintf(&out, "## Chapter %d · Turns %d–%d\n\n", i+1, chapter.FromTurn, chapter.ToTurn)
		fmt.Fprintf(&out, "- Chapter ID: %s\n- Source captured: %s\n- Rules version: %s\n- Rules mode: %s\n- Knowledge boundary: %s\n- Source SHA-256: %s\n\n", chapter.ID, chapter.CreatedAt.Format(time.RFC3339Nano), chapter.RulesVersion, chapter.RulesMode, chapter.KnowledgeBoundary, chapter.SourceHash)
		for _, revision := range chapter.Revisions {
			fmt.Fprintf(&out, "### Revision %d\n\n- Revision ID: %s\n- Request ID: %s\n- Saved at: %s\n- Prompt version: %s\n\n", revision.Revision, revision.ID, revision.RequestID, revision.CreatedAt.Format(time.RFC3339Nano), revision.PromptVersion)
			out.WriteString(revision.Content)
			out.WriteString("\n\n")
		}
		out.WriteString("### Captured source history\n\n")
		for _, turn := range chapter.SourceTurns {
			fmt.Fprintf(&out, "#### Turn %d\n\nSource log: %s\n", turn.TurnNumber, turn.LogID)
			if turn.AttemptID != "" {
				fmt.Fprintf(&out, "\nAction receipt: %s\n", turn.AttemptID)
			}
			fmt.Fprintf(&out, "\n**Player:** %s\n\n%s\n\n", turn.PlayerInput, turn.Narrative)
		}
	}
	fmt.Fprintf(&out, "## Attribution\n\n%s\n\n%s\n", export.Attribution, export.Changes)
	return out.String()
}
