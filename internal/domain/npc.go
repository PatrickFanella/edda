package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type NPC struct {
	ID          uuid.UUID
	CampaignID  uuid.UUID
	Name        string
	Description string
	Personality string
	Disposition int // -100 to 100
	LocationID  *uuid.UUID
	FactionID   *uuid.UUID
	Alive       bool
	HP          *int
	Stats       json.RawMessage
	Properties  json.RawMessage
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (n *NPC) Validate() error {
	if n.Name == "" {
		return errors.New("npc name is required")
	}
	if n.CampaignID == uuid.Nil {
		return errors.New("npc campaign_id is required")
	}
	if n.Disposition < -100 || n.Disposition > 100 {
		return errors.New("npc disposition must be between -100 and 100")
	}
	return nil
}
