package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"github.com/google/uuid"
)

const updateQuestToolName = "update_quest"

// UpdateQuestStore provides quest lookup and persistence for update_quest.
type UpdateQuestStore interface {
	UpdateQuest(ctx context.Context, cmd domain.UpdateQuestMutationCommand) (*domain.UpdateQuestMutationResult, error)
}

// UpdateQuestTool returns the update_quest tool definition and JSON schema.
func UpdateQuestTool() llm.Tool {
	return llm.Tool{
		Name:        updateQuestToolName,
		Description: "Update a quest's status, add objectives, and/or update the description.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"quest_id": map[string]any{
					"type":        "string",
					"description": "Quest UUID.",
				},
				"status": map[string]any{
					"type":        "string",
					"description": "Updated quest status.",
					"enum": []string{
						string(domain.QuestStatusActive),
						string(domain.QuestStatusCompleted),
						string(domain.QuestStatusFailed),
						string(domain.QuestStatusAbandoned),
					},
				},
				"new_objectives": map[string]any{
					"type":        "array",
					"description": "Objective descriptions to append to this quest.",
					"items": map[string]any{
						"type": "string",
					},
				},
				"description_update": map[string]any{
					"type":        "string",
					"description": "Replacement quest description.",
				},
			},
			"required":             []string{"quest_id"},
			"additionalProperties": false,
		},
	}
}

// RegisterUpdateQuest registers the update_quest tool and handler.
func RegisterUpdateQuest(reg *Registry, store UpdateQuestStore) error {
	if store == nil {
		return errors.New("update_quest store is required")
	}
	return reg.Register(UpdateQuestTool(), NewUpdateQuestHandler(store).Handle)
}

// UpdateQuestHandler executes update_quest tool calls.
type UpdateQuestHandler struct {
	store UpdateQuestStore
}

// NewUpdateQuestHandler creates a new update_quest handler.
func NewUpdateQuestHandler(store UpdateQuestStore) *UpdateQuestHandler {
	return &UpdateQuestHandler{store: store}
}

// Handle executes the update_quest tool.
func (h *UpdateQuestHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("update_quest handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("update_quest store is required")
	}
	campaignID, ok := CurrentCampaignIDFromContext(ctx)
	if !ok {
		return nil, errors.New("current campaign context is required")
	}

	questID, err := parseUUIDArg(args, "quest_id")
	if err != nil {
		return nil, err
	}

	statusUpdate, hasStatus, err := parseQuestStatusUpdateArg(args, "status")
	if err != nil {
		return nil, err
	}
	descriptionUpdate, hasDescriptionUpdate, err := parseOptionalQuestDescriptionUpdateArg(args, "description_update")
	if err != nil {
		return nil, err
	}
	newObjectives, hasNewObjectives, err := parseOptionalObjectiveDescriptionsArg(args, "new_objectives")
	if err != nil {
		return nil, err
	}
	if !hasStatus && !hasDescriptionUpdate && !hasNewObjectives {
		return nil, errors.New("at least one of status, description_update, or new_objectives is required")
	}
	cmd := domain.UpdateQuestMutationCommand{CampaignID: campaignID, QuestID: questID, Objectives: newObjectives}
	if hasStatus {
		status := domain.QuestStatus(statusUpdate)
		cmd.Status = &status
	}
	if hasDescriptionUpdate {
		cmd.Description = &descriptionUpdate
	}
	result, err := h.store.UpdateQuest(ctx, cmd)
	if err != nil {
		return nil, err
	}
	addedObjectives := make([]map[string]any, 0, len(result.AddedObjectives))
	for _, objective := range result.AddedObjectives {
		addedObjectives = append(addedObjectives, map[string]any{"id": objective.ID.String(), "description": objective.Description, "completed": objective.Completed, "order_index": objective.OrderIndex})
	}
	objectives := make([]map[string]any, 0, len(result.Objectives))
	for _, objective := range result.Objectives {
		objectives = append(objectives, map[string]any{"id": objective.ID.String(), "description": objective.Description, "completed": objective.Completed, "order_index": objective.OrderIndex})
	}
	cascadedSubquests := make([]map[string]any, 0, len(result.CascadedQuests))
	for _, subquest := range result.CascadedQuests {
		cascadedSubquests = append(cascadedSubquests, map[string]any{"id": subquest.ID.String(), "title": subquest.Title, "old_status": subquest.OldStatus, "new_status": subquest.NewStatus})
	}
	return &ToolResult{Success: true, Data: map[string]any{"id": result.Quest.ID.String(), "campaign_id": result.Quest.CampaignID.String(), "parent_quest_id": optionalUUIDString(result.Quest.ParentQuestID), "title": result.Quest.Title, "description": result.Quest.Description, "quest_type": result.Quest.QuestType, "status": result.Quest.Status, "added_objectives": addedObjectives, "objectives": objectives, "cascaded_subquests": cascadedSubquests}, Narrative: result.Narrative}, nil
}

func parseQuestStatusUpdateArg(args map[string]any, key string) (string, bool, error) {
	if _, ok := args[key]; !ok {
		return "", false, nil
	}
	status, err := parseStringArg(args, key)
	if err != nil {
		return "", false, err
	}
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case string(domain.QuestStatusActive), string(domain.QuestStatusCompleted), string(domain.QuestStatusFailed), string(domain.QuestStatusAbandoned):
		return status, true, nil
	default:
		return "", false, errors.New("status must be one of: active, completed, failed, abandoned")
	}
}

func parseOptionalQuestDescriptionUpdateArg(args map[string]any, key string) (string, bool, error) {
	if _, ok := args[key]; !ok {
		return "", false, nil
	}
	value, err := parseStringArg(args, key)
	if err != nil {
		return "", false, err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false, fmt.Errorf("%s must not be empty or whitespace", key)
	}
	return value, true, nil
}

func parseOptionalObjectiveDescriptionsArg(args map[string]any, key string) ([]string, bool, error) {
	raw, ok := args[key]
	if !ok {
		return nil, false, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, false, fmt.Errorf("%s must be an array", key)
	}
	if len(items) == 0 {
		return nil, false, fmt.Errorf("%s must contain at least one objective", key)
	}
	out := make([]string, 0, len(items))
	for i, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, false, fmt.Errorf("%s[%d] must be a string", key, i)
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, false, fmt.Errorf("%s[%d] must be a non-empty string", key, i)
		}
		out = append(out, value)
	}
	return out, true, nil
}

func optionalUUIDString(id *uuid.UUID) any {
	if id == nil {
		return nil
	}
	return id.String()
}
