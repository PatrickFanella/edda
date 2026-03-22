package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Location struct {
	ID           uuid.UUID
	CampaignID   uuid.UUID
	Name         string
	Description  string
	Region       string
	LocationType string
	Properties   json.RawMessage
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (l *Location) Validate() error {
	if l.Name == "" {
		return errors.New("location name is required")
	}
	if l.CampaignID == uuid.Nil {
		return errors.New("location campaign_id is required")
	}
	return nil
}

type LocationConnection struct {
	ID             uuid.UUID
	FromLocationID uuid.UUID
	ToLocationID   uuid.UUID
	Description    string
	Bidirectional  bool
	TravelTime     string
	CampaignID     uuid.UUID
}
