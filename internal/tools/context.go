package tools

import (
	"context"

	"github.com/google/uuid"
)

type currentLocationIDContextKey struct{}

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
