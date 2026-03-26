package game

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

// sceneStore adapts statedb.Querier to the tools.DescribeSceneStore interface.
type sceneStore struct {
	queries statedb.Querier
}

// NewSceneStore creates a tools.DescribeSceneStore backed by the given Querier.
func NewSceneStore(q statedb.Querier) tools.DescribeSceneStore {
	return &sceneStore{queries: q}
}

func (s *sceneStore) UpdateScene(ctx context.Context, locationID uuid.UUID, description string, mood, timeOfDay *string) error {
	location, err := s.queries.GetLocationByID(ctx, dbutil.ToPgtype(locationID))
	if err != nil {
		return fmt.Errorf("get location: %w", err)
	}

	properties := map[string]any{}
	if len(location.Properties) > 0 {
		if err := json.Unmarshal(location.Properties, &properties); err != nil {
			return fmt.Errorf("unmarshal location properties: %w", err)
		}
	}
	if mood != nil {
		properties["mood"] = *mood
	}
	if timeOfDay != nil {
		properties["time_of_day"] = *timeOfDay
	}
	propertiesJSON, err := json.Marshal(properties)
	if err != nil {
		return fmt.Errorf("marshal location properties: %w", err)
	}

	_, err = s.queries.UpdateLocation(ctx, statedb.UpdateLocationParams{
		ID:           dbutil.ToPgtype(locationID),
		Name:         location.Name,
		Description:  pgtype.Text{String: description, Valid: true},
		Region:       location.Region,
		LocationType: location.LocationType,
		Properties:   propertiesJSON,
	})
	if err != nil {
		return fmt.Errorf("update location: %w", err)
	}

	return nil
}
