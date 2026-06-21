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

var (
	ErrWorldFactNotFound   = errors.New("world fact not found")
	ErrWorldFactSuperseded = errors.New("world fact already superseded")
)

// ReviseWorldFactCommand revises a canonical world fact inside a campaign.
type ReviseWorldFactCommand struct {
	CampaignID     uuid.UUID
	FactID         uuid.UUID
	NewFact        string
	RevealToPlayer bool
}

// ReviseWorldFactResult reports the new fact and propagation details.
type ReviseWorldFactResult struct {
	OldFact               WorldFact
	NewFact               WorldFact
	PlayerKnownPropagated bool
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
