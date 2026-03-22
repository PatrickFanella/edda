package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ItemType string

const (
	ItemTypeWeapon     ItemType = "weapon"
	ItemTypeArmor      ItemType = "armor"
	ItemTypeConsumable ItemType = "consumable"
	ItemTypeQuest      ItemType = "quest"
	ItemTypeMisc       ItemType = "misc"
)

type Item struct {
	ID                uuid.UUID
	CampaignID        uuid.UUID
	PlayerCharacterID *uuid.UUID
	Name              string
	Description       string
	ItemType          ItemType
	Rarity            string
	Properties        json.RawMessage
	Equipped          bool
	Quantity          int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (i *Item) Validate() error {
	if i.Name == "" {
		return errors.New("item name is required")
	}
	if i.CampaignID == uuid.Nil {
		return errors.New("item campaign_id is required")
	}
	if i.Quantity < 0 {
		return errors.New("item quantity cannot be negative")
	}
	switch i.ItemType {
	case ItemTypeWeapon, ItemTypeArmor, ItemTypeConsumable, ItemTypeQuest, ItemTypeMisc:
	default:
		return errors.New("invalid item type")
	}
	return nil
}
