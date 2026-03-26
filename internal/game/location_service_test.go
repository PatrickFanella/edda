package game

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

var _ tools.MovePlayerStore = (*locationService)(nil)
var _ tools.DescribeSceneStore = (*locationService)(nil)

func TestLocationServiceGetLocation(t *testing.T) {
	q := newMockQuerier()
	locID := uuid.New()
	q.location = statedb.Location{
		ID:          dbutil.ToPgtype(locID),
		Name:        "Thornwood Village",
		Description: pgtype.Text{String: "A small village.", Valid: true},
	}
	svc := NewLocationService(q)

	name, desc, err := svc.GetLocation(context.Background(), locID)
	if err != nil {
		t.Fatalf("GetLocation() error = %v", err)
	}
	if name != "Thornwood Village" {
		t.Fatalf("name = %q, want Thornwood Village", name)
	}
	if desc != "A small village." {
		t.Fatalf("description = %q, want A small village.", desc)
	}
}

func TestLocationServiceGetLocationNotFound(t *testing.T) {
	q := newMockQuerier()
	q.getLocationErr = errNotFound
	svc := NewLocationService(q)

	_, _, err := svc.GetLocation(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error for missing location")
	}
}

func TestLocationServiceIsLocationConnectedTrue(t *testing.T) {
	q := newMockQuerier()
	fromID := uuid.New()
	toID := uuid.New()
	campaignID := uuid.New()

	q.location = statedb.Location{
		ID:         dbutil.ToPgtype(fromID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	q.connections = []statedb.GetConnectionsFromLocationRow{
		{ConnectedLocationID: dbutil.ToPgtype(toID)},
	}
	svc := NewLocationService(q)

	connected, err := svc.IsLocationConnected(context.Background(), fromID, toID)
	if err != nil {
		t.Fatalf("IsLocationConnected() error = %v", err)
	}
	if !connected {
		t.Fatal("expected locations to be connected")
	}
}

func TestLocationServiceIsLocationConnectedFalse(t *testing.T) {
	q := newMockQuerier()
	fromID := uuid.New()
	otherID := uuid.New()
	campaignID := uuid.New()

	q.location = statedb.Location{
		ID:         dbutil.ToPgtype(fromID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	q.connections = []statedb.GetConnectionsFromLocationRow{
		{ConnectedLocationID: dbutil.ToPgtype(uuid.New())},
	}
	svc := NewLocationService(q)

	connected, err := svc.IsLocationConnected(context.Background(), fromID, otherID)
	if err != nil {
		t.Fatalf("IsLocationConnected() error = %v", err)
	}
	if connected {
		t.Fatal("expected locations to NOT be connected")
	}
}

func TestLocationServiceUpdatePlayerLocation(t *testing.T) {
	q := newMockQuerier()
	playerID := uuid.New()
	locID := uuid.New()
	svc := NewLocationService(q)

	err := svc.UpdatePlayerLocation(context.Background(), playerID, locID)
	if err != nil {
		t.Fatalf("UpdatePlayerLocation() error = %v", err)
	}
	if q.lastUpdatePlayerLocParams == nil {
		t.Fatal("expected UpdatePlayerLocation to be called")
	}
	if dbutil.FromPgtype(q.lastUpdatePlayerLocParams.ID) != playerID {
		t.Fatalf("player ID = %v, want %s", q.lastUpdatePlayerLocParams.ID, playerID)
	}
	if dbutil.FromPgtype(q.lastUpdatePlayerLocParams.CurrentLocationID) != locID {
		t.Fatalf("location ID = %v, want %s", q.lastUpdatePlayerLocParams.CurrentLocationID, locID)
	}
}

func TestLocationServiceUpdateSceneMergesProperties(t *testing.T) {
	q := newMockQuerier()
	locID := uuid.New()
	q.location = statedb.Location{
		ID:         dbutil.ToPgtype(locID),
		Name:       "Town Square",
		Properties: []byte(`{"weather":"rain","mood":"calm"}`),
	}
	svc := NewLocationService(q)

	mood := "tense"
	timeOfDay := "dusk"
	err := svc.UpdateScene(context.Background(), locID, "A tense square.", &mood, &timeOfDay)
	if err != nil {
		t.Fatalf("UpdateScene() error = %v", err)
	}
	if q.lastUpdateLocationParams == nil {
		t.Fatal("expected UpdateLocation to be called")
	}

	var props map[string]any
	if err := json.Unmarshal(q.lastUpdateLocationParams.Properties, &props); err != nil {
		t.Fatalf("unmarshal properties: %v", err)
	}
	if props["weather"] != "rain" {
		t.Fatalf("existing key 'weather' lost: %v", props)
	}
	if props["mood"] != "tense" {
		t.Fatalf("mood not updated: %v", props)
	}
	if props["time_of_day"] != "dusk" {
		t.Fatalf("time_of_day not added: %v", props)
	}
}

func TestLocationServiceUpdateSceneNilOptionals(t *testing.T) {
	q := newMockQuerier()
	locID := uuid.New()
	q.location = statedb.Location{
		ID:         dbutil.ToPgtype(locID),
		Name:       "Town Square",
		Properties: []byte(`{"mood":"calm"}`),
	}
	svc := NewLocationService(q)

	err := svc.UpdateScene(context.Background(), locID, "Updated description.", nil, nil)
	if err != nil {
		t.Fatalf("UpdateScene() error = %v", err)
	}

	var props map[string]any
	if err := json.Unmarshal(q.lastUpdateLocationParams.Properties, &props); err != nil {
		t.Fatalf("unmarshal properties: %v", err)
	}
	if props["mood"] != "calm" {
		t.Fatalf("mood should be preserved when nil: %v", props)
	}
}

var errNotFound = pgx.ErrNoRows
