package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorldFact struct {
	ID           uuid.UUID
	CampaignID   uuid.UUID
	Fact         string
	Category     string
	Source       string
	SupersededBy *uuid.UUID
	CreatedAt    time.Time
}

func (wf *WorldFact) Validate() error {
	if wf.Fact == "" {
		return errors.New("world fact text is required")
	}
	if wf.CampaignID == uuid.Nil {
		return errors.New("world fact campaign_id is required")
	}
	return nil
}
