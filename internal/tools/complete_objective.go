package tools

import (
	"context"
	"errors"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"github.com/google/uuid"
)

const completeObjectiveToolName = "complete_objective"

// CompleteObjectiveStore provides quest objective lookup and persistence for complete_objective.
type CompleteObjectiveStore interface {
	CompleteObjective(ctx context.Context, cmd domain.CompleteObjectiveMutationCommand) (*domain.CompleteObjectiveMutationResult, error)
}

// CompleteObjectiveTool returns the complete_objective tool definition and JSON schema.
func CompleteObjectiveTool() llm.Tool {
	return llm.Tool{
		Name:        completeObjectiveToolName,
		Description: "Complete a single quest objective by objective ID or objective description match.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"quest_id": map[string]any{
					"type":        "string",
					"description": "Quest UUID.",
				},
				"objective_id": map[string]any{
					"type":        "string",
					"description": "Objective UUID to complete.",
				},
				"objective_description": map[string]any{
					"type":        "string",
					"description": "Objective description text to match when objective_id is unknown.",
				},
				"auto_complete_quest": map[string]any{
					"type":        "boolean",
					"description": "When true, automatically mark quest completed once all objectives are complete.",
				},
			},
			"required":             []string{"quest_id"},
			"additionalProperties": false,
		},
	}
}

// RegisterCompleteObjective registers the complete_objective tool and handler.
func RegisterCompleteObjective(reg *Registry, store CompleteObjectiveStore) error {
	if store == nil {
		return errors.New("complete_objective store is required")
	}
	return reg.Register(CompleteObjectiveTool(), NewCompleteObjectiveHandler(store).Handle)
}

// CompleteObjectiveHandler executes complete_objective tool calls.
type CompleteObjectiveHandler struct {
	store CompleteObjectiveStore
}

// NewCompleteObjectiveHandler creates a new complete_objective handler.
func NewCompleteObjectiveHandler(store CompleteObjectiveStore) *CompleteObjectiveHandler {
	return &CompleteObjectiveHandler{store: store}
}

// Handle executes the complete_objective tool.
func (h *CompleteObjectiveHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("complete_objective handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("complete_objective store is required")
	}
	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("current campaign context is required")
	}

	questID, err := parseUUIDArg(args, "quest_id")
	if err != nil {
		return nil, err
	}

	var objectiveID uuid.UUID
	objectiveIDSet := false
	if _, ok := args["objective_id"]; ok {
		objectiveID, err = parseUUIDArg(args, "objective_id")
		if err != nil {
			return nil, err
		}
		objectiveIDSet = true
	}

	var objectiveDescription string
	descriptionSet := false
	if _, ok := args["objective_description"]; ok {
		objectiveDescription, err = parseStringArg(args, "objective_description")
		if err != nil {
			return nil, err
		}
		objectiveDescription = strings.TrimSpace(objectiveDescription)
		if objectiveDescription == "" {
			return nil, errors.New("objective_description must not be empty or whitespace")
		}
		descriptionSet = true
	}

	if objectiveIDSet == descriptionSet {
		return nil, errors.New("exactly one of objective_id or objective_description must be provided")
	}
	var objectiveIDPtr *uuid.UUID
	if objectiveIDSet {
		objectiveIDPtr = &objectiveID
	}
	cmd := domain.CompleteObjectiveMutationCommand{CampaignID: campaignID, QuestID: questID, ObjectiveID: objectiveIDPtr, ObjectiveDescription: &objectiveDescription}
	if !descriptionSet {
		cmd.ObjectiveDescription = nil
	}
	if !objectiveIDSet {
		cmd.ObjectiveID = nil
	}
	if _, ok := args["auto_complete_quest"]; ok {
		cmd.AutoCompleteQuest, err = parseBoolArg(args, "auto_complete_quest")
		if err != nil {
			return nil, err
		}
	}
	result, err := h.store.CompleteObjective(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return &ToolResult{Success: true, Data: map[string]any{"quest_id": result.Quest.ID.String(), "quest_title": result.Quest.Title, "quest_status": result.Quest.Status, "objective_id": result.Objective.ID.String(), "objective_description": result.Objective.Description, "objective_completed": true, "objectives_completed": result.ObjectivesCompleted, "objectives_total": result.ObjectivesTotal, "progress": result.Progress, "all_objectives_complete": result.AllObjectivesComplete, "quest_ready_for_completion": result.QuestReadyForCompletion, "quest_auto_completed": result.QuestAutoCompleted}, Narrative: result.Narrative}, nil
}
