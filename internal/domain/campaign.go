package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type CampaignStatus string

const (
	CampaignStatusActive    CampaignStatus = "active"
	CampaignStatusPaused    CampaignStatus = "paused"
	CampaignStatusCompleted CampaignStatus = "completed"
)

type Campaign struct {
	ID          uuid.UUID
	Name        string
	Description string
	Genre       string
	Tone        string
	Themes      []string
	Status      CampaignStatus
	CreatedBy   uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (c *Campaign) Validate() error {
	if c.Name == "" {
		return errors.New("campaign name is required")
	}
	switch c.Status {
	case CampaignStatusActive, CampaignStatusPaused, CampaignStatusCompleted:
	default:
		return errors.New("campaign status must be active, paused, or completed")
	}
	if c.CreatedBy == uuid.Nil {
		return errors.New("campaign created_by is required")
	}
	return nil
}
