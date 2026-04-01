package tools

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/llm"
)

const describeSceneToolName = "describe_scene"

// DescribeSceneStore persists scene description updates for a location.
type DescribeSceneStore interface {
	UpdateScene(ctx context.Context, locationID uuid.UUID, description string, mood, timeOfDay *string) error
}

// DescribeSceneTool returns the describe_scene tool definition and JSON schema.
func DescribeSceneTool() llm.Tool {
	return llm.Tool{
		Name:        describeSceneToolName,
		Description: "Set or update the narrative scene description for the current location.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"description": map[string]any{
					"type":        "string",
					"description": "The current scene description to present to the player.",
				},
				"mood": map[string]any{
					"type":        "string",
					"description": "Optional scene mood for the location properties.",
				},
				"time_of_day": map[string]any{
					"type":        "string",
					"description": "Optional time of day for the location properties.",
				},
			},
			"required":             []string{"description"},
			"additionalProperties": false,
		},
	}
}

// RegisterDescribeScene registers the describe_scene tool and handler.
func RegisterDescribeScene(reg *Registry, store DescribeSceneStore) error {
	if store == nil {
		return errors.New("describe_scene store is required")
	}
	return reg.Register(DescribeSceneTool(), NewDescribeSceneHandler(store).Handle)
}

// DescribeSceneHandler executes describe_scene tool calls.
type DescribeSceneHandler struct {
	store DescribeSceneStore
}

// NewDescribeSceneHandler creates a new describe_scene handler.
func NewDescribeSceneHandler(store DescribeSceneStore) *DescribeSceneHandler {
	return &DescribeSceneHandler{store: store}
}

// Handle executes the describe_scene tool.
func (h *DescribeSceneHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("describe_scene handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("describe_scene store is required")
	}

	description, err := parseStringArg(args, "description")
	if err != nil {
		return nil, err
	}

	mood, err := parseOptionalNonEmptyStringArg(args, "mood")
	if err != nil {
		return nil, err
	}
	timeOfDay, err := parseOptionalNonEmptyStringArg(args, "time_of_day")
	if err != nil {
		return nil, err
	}

	locationID, ok := CurrentLocationIDFromContext(ctx)
	if !ok {
		return nil, errors.New("describe_scene requires current location id in context")
	}

	if err := h.store.UpdateScene(ctx, locationID, description, mood, timeOfDay); err != nil {
		return nil, fmt.Errorf("update scene: %w", err)
	}

	data := map[string]any{
		"location_id": locationID.String(),
		"description": description,
	}
	if mood != nil {
		data["mood"] = *mood
	}
	if timeOfDay != nil {
		data["time_of_day"] = *timeOfDay
	}

	return &ToolResult{
		Success:   true,
		Data:      data,
		Narrative: "Scene description updated successfully.",
	}, nil
}
