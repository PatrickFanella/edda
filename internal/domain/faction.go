package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Faction struct {
	ID          uuid.UUID
	CampaignID  uuid.UUID
	Name        string
	Description string
	Agenda      string
	Territory   string
	Properties  json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (f *Faction) Validate() error {
	if f.Name == "" {
		return errors.New("faction name is required")
	}
	if f.CampaignID == uuid.Nil {
		return errors.New("faction campaign_id is required")
	}
	return nil
}

type FactionRelationship struct {
	ID               uuid.UUID
	FactionID        uuid.UUID
	RelatedFactionID uuid.UUID
	RelationshipType string
	Description      string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
