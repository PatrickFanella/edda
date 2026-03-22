package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type MemoryType string

const (
	MemoryTypeTurnSummary    MemoryType = "turn_summary"
	MemoryTypeLore           MemoryType = "lore"
	MemoryTypeNPCInteraction MemoryType = "npc_interaction"
	MemoryTypeScene          MemoryType = "scene"
	MemoryTypeWorldFact      MemoryType = "world_fact"
)

type Memory struct {
	ID           uuid.UUID
	CampaignID   uuid.UUID
	Content      string
	Embedding    []float32
	MemoryType   MemoryType
	LocationID   *uuid.UUID
	NPCsInvolved []uuid.UUID
	InGameTime   string
	Metadata     json.RawMessage
	CreatedAt    time.Time
}

func (m *Memory) Validate() error {
	if m.CampaignID == uuid.Nil {
		return errors.New("memory campaign_id is required")
	}
	if m.Content == "" {
		return errors.New("memory content is required")
	}
	return nil
}
