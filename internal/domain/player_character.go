package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PlayerCharacter struct {
	ID                uuid.UUID
	CampaignID        uuid.UUID
	UserID            uuid.UUID
	Name              string
	Description       string
	Stats             json.RawMessage
	HP                int
	MaxHP             int
	Experience        int
	Level             int
	Status            string
	Abilities         json.RawMessage
	CurrentLocationID *uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (pc *PlayerCharacter) Validate() error {
	if pc.Name == "" {
		return errors.New("player character name is required")
	}
	if pc.CampaignID == uuid.Nil {
		return errors.New("player character campaign_id is required")
	}
	if pc.UserID == uuid.Nil {
		return errors.New("player character user_id is required")
	}
	if pc.Level < 1 {
		return errors.New("player character level must be at least 1")
	}
	return nil
}
