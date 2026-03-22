package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type EntityType string

const (
	EntityTypeNPC             EntityType = "npc"
	EntityTypeLocation        EntityType = "location"
	EntityTypeFaction         EntityType = "faction"
	EntityTypePlayerCharacter EntityType = "player_character"
	EntityTypeItem            EntityType = "item"
)

type EntityRelationship struct {
	ID               uuid.UUID
	CampaignID       uuid.UUID
	SourceEntityType EntityType
	SourceEntityID   uuid.UUID
	TargetEntityType EntityType
	TargetEntityID   uuid.UUID
	RelationshipType string
	Description      string
	Strength         *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (er *EntityRelationship) Validate() error {
	if er.CampaignID == uuid.Nil {
		return errors.New("entity relationship campaign_id is required")
	}
	if er.SourceEntityID == uuid.Nil {
		return errors.New("entity relationship source_entity_id is required")
	}
	if er.TargetEntityID == uuid.Nil {
		return errors.New("entity relationship target_entity_id is required")
	}
	if er.RelationshipType == "" {
		return errors.New("entity relationship type is required")
	}
	return nil
}
