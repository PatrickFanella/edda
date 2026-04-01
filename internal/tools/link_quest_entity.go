package tools

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/llm"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

const linkQuestEntityToolName = "link_quest_entity"

// LinkQuestEntityStore validates quests and entities, then creates the linking relationship.
type LinkQuestEntityStore interface {
	GetLocationByID(ctx context.Context, id pgtype.UUID) (statedb.Location, error)
	GetQuestByID(ctx context.Context, id pgtype.UUID) (statedb.Quest, error)
	GetNPCByID(ctx context.Context, id pgtype.UUID) (statedb.Npc, error)
	GetFactionByID(ctx context.Context, id pgtype.UUID) (statedb.Faction, error)
	GetPlayerCharacterByID(ctx context.Context, id pgtype.UUID) (statedb.PlayerCharacter, error)
	GetItemByID(ctx context.Context, id pgtype.UUID) (statedb.Item, error)
	CreateRelationship(ctx context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error)
	GetRelationshipsByEntity(ctx context.Context, arg statedb.GetRelationshipsByEntityParams) ([]statedb.EntityRelationship, error)
}

// LinkQuestEntityTool returns the link_quest_entity tool definition and JSON schema.
func LinkQuestEntityTool() llm.Tool {
	return llm.Tool{
		Name:        linkQuestEntityToolName,
		Description: "Link an existing entity (NPC, location, faction, player character, or item) to a quest.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"quest_id": map[string]any{
					"type":        "string",
					"description": "UUID of the quest to link.",
				},
				"entity_type": map[string]any{
					"type":        "string",
					"description": "Entity type (npc, location, faction, player_character, item). You may also provide player as an alias for player_character.",
				},
				"entity_id": map[string]any{
					"type":        "string",
					"description": "UUID of the entity to link.",
				},
				"link_description": map[string]any{
					"type":        "string",
					"description": "Narrative description of how this entity relates to the quest.",
				},
				"link_type": map[string]any{
					"type":        "string",
					"description": "Optional relationship category. Defaults to quest_related.",
				},
			},
			"required":             []string{"quest_id", "entity_type", "entity_id", "link_description"},
			"additionalProperties": false,
		},
	}
}

// RegisterLinkQuestEntity registers the link_quest_entity tool and handler.
func RegisterLinkQuestEntity(reg *Registry, store LinkQuestEntityStore) error {
	if store == nil {
		return errors.New("link_quest_entity store is required")
	}
	return reg.Register(LinkQuestEntityTool(), NewLinkQuestEntityHandler(store).Handle)
}

// LinkQuestEntityHandler executes link_quest_entity tool calls.
type LinkQuestEntityHandler struct {
	store LinkQuestEntityStore
}

// NewLinkQuestEntityHandler creates a new link_quest_entity handler.
func NewLinkQuestEntityHandler(store LinkQuestEntityStore) *LinkQuestEntityHandler {
	return &LinkQuestEntityHandler{store: store}
}

// Handle executes the link_quest_entity tool.
func (h *LinkQuestEntityHandler) Handle(ctx context.Context, args map[string]any) (*ToolResult, error) {
	if h == nil {
		return nil, errors.New("link_quest_entity handler is nil")
	}
	if h.store == nil {
		return nil, errors.New("link_quest_entity store is required")
	}

	questID, err := parseUUIDArg(args, "quest_id")
	if err != nil {
		return nil, err
	}
	entityType, err := parseRelationshipEntityTypeArg(args, "entity_type")
	if err != nil {
		return nil, err
	}
	entityID, err := parseUUIDArg(args, "entity_id")
	if err != nil {
		return nil, err
	}
	linkDescription, err := parseStringArg(args, "link_description")
	if err != nil {
		return nil, err
	}
	trimmedDescription := strings.TrimSpace(linkDescription)
	if trimmedDescription == "" {
		return nil, errors.New("link_description must not be empty or whitespace")
	}

	linkType := "quest_related"
	if raw, ok := args["link_type"]; ok {
		if s, ok := raw.(string); ok {
			trimmed := strings.TrimSpace(s)
			if trimmed != "" {
				linkType = trimmed
			}
		}
	}

	// Validate quest exists.
	quest, err := h.store.GetQuestByID(ctx, dbutil.ToPgtype(questID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("quest not found: %s", questID)
		}
		return nil, fmt.Errorf("lookup quest: %w", err)
	}

	// Resolve campaign from current location.
	currentLocationID, ok := CurrentLocationIDFromContext(ctx)
	if !ok {
		return nil, errors.New("link_quest_entity requires current location id in context")
	}
	currentLocation, err := h.store.GetLocationByID(ctx, dbutil.ToPgtype(currentLocationID))
	if err != nil {
		return nil, fmt.Errorf("resolve campaign from current location: %w", err)
	}
	campaignID := dbutil.FromPgtype(currentLocation.CampaignID)

	// Validate entity exists and belongs to the same campaign.
	if err := h.validateEntityInCampaign(ctx, campaignID, entityType, entityID); err != nil {
		return nil, err
	}

	relationship, err := h.store.CreateRelationship(ctx, statedb.CreateRelationshipParams{
		CampaignID:       dbutil.ToPgtype(campaignID),
		SourceEntityType: "quest",
		SourceEntityID:   quest.ID,
		TargetEntityType: entityType,
		TargetEntityID:   dbutil.ToPgtype(entityID),
		RelationshipType: linkType,
		Description:      pgtype.Text{String: trimmedDescription, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create quest-entity link: %w", err)
	}

	summary := fmt.Sprintf("Linked %s %s to quest %q via %q.",
		entityType, entityID.String(), quest.Title, linkType)

	return &ToolResult{
		Success: true,
		Data: map[string]any{
			"id":                 dbutil.FromPgtype(relationship.ID).String(),
			"campaign_id":        campaignID.String(),
			"quest_id":           questID.String(),
			"source_entity_type": relationship.SourceEntityType,
			"source_entity_id":   dbutil.FromPgtype(relationship.SourceEntityID).String(),
			"target_entity_type": relationship.TargetEntityType,
			"target_entity_id":   dbutil.FromPgtype(relationship.TargetEntityID).String(),
			"relationship_type":  relationship.RelationshipType,
			"description":        relationship.Description.String,
			"summary":            summary,
		},
		Narrative: summary,
	}, nil
}

func (h *LinkQuestEntityHandler) validateEntityInCampaign(ctx context.Context, campaignID uuid.UUID, entityType string, entityID uuid.UUID) error {
	switch domain.EntityType(entityType) {
	case domain.EntityTypeNPC:
		entity, err := h.store.GetNPCByID(ctx, dbutil.ToPgtype(entityID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("entity not found: npc %s", entityID)
			}
			return fmt.Errorf("lookup npc: %w", err)
		}
		if dbutil.FromPgtype(entity.CampaignID) != campaignID {
			return fmt.Errorf("entity does not belong to current campaign: npc %s", entityID)
		}
	case domain.EntityTypeLocation:
		entity, err := h.store.GetLocationByID(ctx, dbutil.ToPgtype(entityID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("entity not found: location %s", entityID)
			}
			return fmt.Errorf("lookup location: %w", err)
		}
		if dbutil.FromPgtype(entity.CampaignID) != campaignID {
			return fmt.Errorf("entity does not belong to current campaign: location %s", entityID)
		}
	case domain.EntityTypeFaction:
		entity, err := h.store.GetFactionByID(ctx, dbutil.ToPgtype(entityID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("entity not found: faction %s", entityID)
			}
			return fmt.Errorf("lookup faction: %w", err)
		}
		if dbutil.FromPgtype(entity.CampaignID) != campaignID {
			return fmt.Errorf("entity does not belong to current campaign: faction %s", entityID)
		}
	case domain.EntityTypePlayerCharacter:
		entity, err := h.store.GetPlayerCharacterByID(ctx, dbutil.ToPgtype(entityID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("entity not found: player_character %s", entityID)
			}
			return fmt.Errorf("lookup player character: %w", err)
		}
		if dbutil.FromPgtype(entity.CampaignID) != campaignID {
			return fmt.Errorf("entity does not belong to current campaign: player_character %s", entityID)
		}
	case domain.EntityTypeItem:
		entity, err := h.store.GetItemByID(ctx, dbutil.ToPgtype(entityID))
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("entity not found: item %s", entityID)
			}
			return fmt.Errorf("lookup item: %w", err)
		}
		if dbutil.FromPgtype(entity.CampaignID) != campaignID {
			return fmt.Errorf("entity does not belong to current campaign: item %s", entityID)
		}
	default:
		return fmt.Errorf("entity_type must be one of: npc, location, faction, player_character, player, item")
	}
	return nil
}
