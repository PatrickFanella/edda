package domain

import "github.com/google/uuid"

// RevealLocationCommand reveals a location inside a campaign.
type RevealLocationCommand struct {
	CampaignID uuid.UUID
	LocationID uuid.UUID
}

// RevealLocationResult reports the revealed location.
type RevealLocationResult struct {
	LocationID   uuid.UUID
	LocationName string
}

// MovePlayerCommand moves a player character between campaign locations.
type MovePlayerCommand struct {
	CampaignID        uuid.UUID
	PlayerCharacterID uuid.UUID
	CurrentLocationID uuid.UUID
	TargetLocationID  uuid.UUID
}

// MovePlayerResult reports the move outcome and any tolerated warnings.
type MovePlayerResult struct {
	PlayerCharacterID     uuid.UUID
	FromLocationID        uuid.UUID
	ToLocationID          uuid.UUID
	ToLocationName        string
	ToLocationDescription string
	TravelTime            string
	Day                   int
	Hour                  int
	Minute                int
	VisitedMarked         bool
	VisitedWarning        string
	TimeWarning           string
}
