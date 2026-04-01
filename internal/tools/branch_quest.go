package tools

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/llm"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

const branchQuestToolName = "branch_quest"

// BranchQuestStore persists and queries quests, objectives, and relationships for branching.
type BranchQuestStore interface {
	GetLocationByID(ctx context.Context, id pgtype.UUID) (statedb.Location, error)
	GetQuestByID(ctx context.Context, id pgtype.UUID) (statedb.Quest, error)
	CreateQuest(ctx context.Context, arg statedb.CreateQuestParams) (statedb.Quest, error)
	CreateObjective(ctx context.Context, arg statedb.CreateObjectiveParams) (statedb.QuestObjective, error)
	CreateRelationship(ctx context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error)
	UpdateQuestStatus(ctx context.Context, arg statedb.UpdateQuestStatusParams) (statedb.Quest, error)
	ListObjectivesByQuest(ctx context.Context, questID pgtype.UUID) ([]statedb.QuestObjective, error)
	GetRelationshipsByEntity(ctx context.Context, arg statedb.GetRelationshipsByEntityParams) ([]statedb.EntityRelationship, error)
}

// BranchQuestTool returns the branch_quest tool definition and JSON schema.
func BranchQuestTool() llm.Tool {
	return llm.Tool{
		Name:        branchQuestToolName,
		Description: "Branch a quest into a new quest path, optionally updating the original quest's status.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"original_quest_id": map[string]any{
					"type":        "string",
					"description": "UUID of the quest to branch from.",
				},
				"branch_title": map[string]any{
					"type":        "string",
					"description": "Title for the new branch quest.",
				},
				"branch_description": map[string]any{
					"type":        "string",
					"description": "Description for the new branch quest.",
				},
				"branch_objectives": map[string]any{
					"type":        "array",
					"description": "Ordered objectives for the branch quest.",
					"items": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"description": map[string]any{
								"type":        "string",
								"description": "Objective description.",
							},
							"order_index": map[string]any{
								"type":        "integer",
								"description": "Objective order within the quest.",
							},
						},
						"required":             []string{"description", "order_index"},
						"additionalProperties": false,
					},
				},
				"reason": map[string]any{
					"type":        "string",
					"description": "What choice or event caused this branch.",
				},
				"original_quest_status": map[string]any{
					"type":        "string",
					"description": "Optional status to set on the original quest after branching.",
					"enum":        []string{string(domain.QuestStatusActive), string(domain.QuestStatusCompleted), string(domain.QuestStatusFailed)},
				},
			},
			"required":             []string{"original_quest_id", "branch_title", "branch_description", "branch_objectives", "reason"},
			"additionalProperties": false,
		},
	}
}

// RegisterBranchQuest registers the branch_quest tool and handler.
func RegisterBranchQuest(reg *Registry, store BranchQuestStore) error {
	if store == nil {
		return errors.New("branch_quest store is required")
	}
	return reg.Register(BranchQuestTool(), NewBranchQuestHandler(store).Handle)
}

// BranchQuestHandler executes branch_quest tool calls.
type BranchQuestHandler struct {
	store BranchQuestStore
}

// NewBranchQuestHandler creates a new branch_quest handler.
func NewBranchQuestHandler(store BranchQuestStore) *BranchQuestHandler {
	return &BranchQuestHandler{store: store}
}

// Handle executes the branch_quest tool.
func (h *BranchQuestHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("branch_quest handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("branch_quest store is required")
	}

	originalQuestID, err := parseUUIDArg(args, "original_quest_id")
	if err != nil {
		return nil, err
	}
	branchTitle, err := parseStringArg(args, "branch_title")
	if err != nil {
		return nil, err
	}
	branchDescription, err := parseStringArg(args, "branch_description")
	if err != nil {
		return nil, err
	}
	objectives, err := parseQuestObjectivesArg(args, "branch_objectives")
	if err != nil {
		return nil, err
	}
	reason, err := parseStringArg(args, "reason")
	if err != nil {
		return nil, err
	}
	originalStatusUpdate, hasOriginalStatus, err := parseQuestStatusUpdateArg(args, "original_quest_status")
	if err != nil {
		return nil, err
	}

	// Look up original quest.
	originalQuest, err := h.store.GetQuestByID(ctx, dbutil.ToPgtype(originalQuestID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("original_quest_id does not reference an existing quest")
		}
		return nil, fmt.Errorf("get original quest: %w", err)
	}

	// Resolve campaign from current location.
	currentLocationID, ok := CurrentLocationIDFromContext(ctx)
	if !ok {
		return nil, errors.New("branch_quest requires current location id in context")
	}
	currentLocation, err := h.store.GetLocationByID(ctx, dbutil.ToPgtype(currentLocationID))
	if err != nil {
		return nil, fmt.Errorf("resolve campaign from current location: %w", err)
	}

	// Create branch quest inheriting quest_type from original.
	branchQuest, err := h.store.CreateQuest(ctx, statedb.CreateQuestParams{
		CampaignID:  currentLocation.CampaignID,
		Title:       strings.TrimSpace(branchTitle),
		Description: pgtype.Text{String: strings.TrimSpace(branchDescription), Valid: true},
		QuestType:   originalQuest.QuestType,
		Status:      string(domain.QuestStatusActive),
	})
	if err != nil {
		return nil, fmt.Errorf("create branch quest: %w", err)
	}

	// Create objectives for the branch.
	slices.SortFunc(objectives, func(a, b createQuestObjectiveInput) int {
		return a.OrderIndex - b.OrderIndex
	})

	createdObjectives := make([]map[string]any, 0, len(objectives))
	for i, objective := range objectives {
		createdObj, err := h.store.CreateObjective(ctx, statedb.CreateObjectiveParams{
			QuestID:     branchQuest.ID,
			Description: objective.Description,
			Completed:   false,
			OrderIndex:  int32(objective.OrderIndex),
		})
		if err != nil {
			return nil, fmt.Errorf("create branch_objectives[%d]: %w", i, err)
		}
		createdObjectives = append(createdObjectives, map[string]any{
			"id":          dbutil.FromPgtype(createdObj.ID).String(),
			"description": createdObj.Description,
			"completed":   createdObj.Completed,
			"order_index": createdObj.OrderIndex,
		})
	}

	// Link branch to original via a quest_branch relationship.
	branchQuestID := dbutil.FromPgtype(branchQuest.ID)
	_, err = h.store.CreateRelationship(ctx, statedb.CreateRelationshipParams{
		CampaignID:       currentLocation.CampaignID,
		SourceEntityType: "quest",
		SourceEntityID:   branchQuest.ID,
		TargetEntityType: "quest",
		TargetEntityID:   dbutil.ToPgtype(originalQuestID),
		RelationshipType: "quest_branch",
		Description:      pgtype.Text{String: fmt.Sprintf("Branched from %q: %s", originalQuest.Title, strings.TrimSpace(reason)), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create branch relationship: %w", err)
	}

	// Copy entity relationships from the original quest to the branch.
	originalRels, err := h.store.GetRelationshipsByEntity(ctx, statedb.GetRelationshipsByEntityParams{
		CampaignID: currentLocation.CampaignID,
		EntityType: "quest",
		EntityID:   dbutil.ToPgtype(originalQuestID),
	})
	if err != nil {
		return nil, fmt.Errorf("get original quest relationships: %w", err)
	}

	copiedCount := 0
	for _, rel := range originalRels {
		// Copy each relationship, substituting the branch quest for the original.
		params := statedb.CreateRelationshipParams{
			CampaignID:       rel.CampaignID,
			SourceEntityType: rel.SourceEntityType,
			SourceEntityID:   rel.SourceEntityID,
			TargetEntityType: rel.TargetEntityType,
			TargetEntityID:   rel.TargetEntityID,
			RelationshipType: rel.RelationshipType,
			Description:      rel.Description,
		}
		// Replace the original quest reference with the branch quest.
		if rel.SourceEntityType == "quest" && rel.SourceEntityID == dbutil.ToPgtype(originalQuestID) {
			params.SourceEntityID = branchQuest.ID
		}
		if rel.TargetEntityType == "quest" && rel.TargetEntityID == dbutil.ToPgtype(originalQuestID) {
			params.TargetEntityID = branchQuest.ID
		}
		if _, err := h.store.CreateRelationship(ctx, params); err != nil {
			return nil, fmt.Errorf("copy relationship: %w", err)
		}
		copiedCount++
	}

	// Optionally update original quest status.
	var originalStatusChange map[string]any
	if hasOriginalStatus {
		updatedOriginal, err := h.store.UpdateQuestStatus(ctx, statedb.UpdateQuestStatusParams{
			Status: originalStatusUpdate,
			ID:     dbutil.ToPgtype(originalQuestID),
		})
		if err != nil {
			return nil, fmt.Errorf("update original quest status: %w", err)
		}
		originalStatusChange = map[string]any{
			"original_quest_id": originalQuestID.String(),
			"old_status":        originalQuest.Status,
			"new_status":        updatedOriginal.Status,
		}
	}

	summary := fmt.Sprintf("Quest %q branched from %q with %d objective(s) and %d copied relationship(s).",
		branchQuest.Title, originalQuest.Title, len(createdObjectives), copiedCount)

	data := map[string]any{
		"id":                    branchQuestID.String(),
		"campaign_id":           dbutil.FromPgtype(branchQuest.CampaignID).String(),
		"title":                 branchQuest.Title,
		"description":           branchQuest.Description.String,
		"quest_type":            branchQuest.QuestType,
		"status":                branchQuest.Status,
		"original_quest_id":     originalQuestID.String(),
		"objectives":            createdObjectives,
		"copied_relationships":  copiedCount,
		"summary":               summary,
	}
	if originalStatusChange != nil {
		data["original_status_change"] = originalStatusChange
	}

	return &ToolResult{
		Success:   true,
		Data:      data,
		Narrative: summary,
	}, nil
}
