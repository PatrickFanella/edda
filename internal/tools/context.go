package tools

import (
	"context"

	"github.com/google/uuid"
)

type currentCampaignIDContextKey struct{}
type currentLocationIDContextKey struct{}
type currentPlayerCharacterIDContextKey struct{}

// WithCurrentCampaignID returns a context carrying the current campaign ID.
func WithCurrentCampaignID(ctx context.Context, campaignID uuid.UUID) context.Context {
	return context.WithValue(ctx, currentCampaignIDContextKey{}, campaignID)
}

// CurrentCampaignIDFromContext returns the current campaign ID from context.
func CurrentCampaignIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	campaignID, ok := ctx.Value(currentCampaignIDContextKey{}).(uuid.UUID)
	if !ok || campaignID == uuid.Nil {
		return uuid.Nil, false
	}
	return campaignID, true
}

// WithCurrentLocationID returns a context carrying the current location ID.
func WithCurrentLocationID(ctx context.Context, locationID uuid.UUID) context.Context {
	return context.WithValue(ctx, currentLocationIDContextKey{}, locationID)
}

// CurrentLocationIDFromContext returns the current location ID from context.
func CurrentLocationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	locationID, ok := ctx.Value(currentLocationIDContextKey{}).(uuid.UUID)
	if !ok || locationID == uuid.Nil {
		return uuid.Nil, false
	}
	return locationID, true
}

// WithCurrentPlayerCharacterID returns a context carrying the current player character ID.
func WithCurrentPlayerCharacterID(ctx context.Context, playerCharacterID uuid.UUID) context.Context {
	return context.WithValue(ctx, currentPlayerCharacterIDContextKey{}, playerCharacterID)
}

// CurrentPlayerCharacterIDFromContext returns the current player character ID from context.
func CurrentPlayerCharacterIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if ctx == nil {
		return uuid.Nil, false
	}
	playerCharacterID, ok := ctx.Value(currentPlayerCharacterIDContextKey{}).(uuid.UUID)
	if !ok || playerCharacterID == uuid.Nil {
		return uuid.Nil, false
	}
	return playerCharacterID, true
}
