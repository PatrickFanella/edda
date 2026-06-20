package game

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

type stateFactStoreStub struct {
	fact           statedb.WorldFact
	newFact        statedb.WorldFact
	location       statedb.Location
	player         statedb.PlayerCharacter
	connections    []statedb.GetConnectionsFromLocationRow
	getErr         error
	getLocationErr error
	connectionsErr error
	supersedeErr   error
	updateErr      error
	knownErr       error
	visitedErr     error
	lastSupersede  statedb.SupersedeFactParams
	lastUpdate     statedb.UpdatePlayerLocationParams
}

func (s *stateFactStoreStub) GetFactByID(_ context.Context, _ pgtype.UUID) (statedb.WorldFact, error) {
	return s.fact, s.getErr
}

func (s *stateFactStoreStub) SupersedeFact(_ context.Context, arg statedb.SupersedeFactParams) (statedb.WorldFact, error) {
	s.lastSupersede = arg
	if s.supersedeErr != nil {
		return statedb.WorldFact{}, s.supersedeErr
	}
	return s.newFact, nil
}

func (s *stateFactStoreStub) GetLocationByID(_ context.Context, _ statedb.GetLocationByIDParams) (statedb.Location, error) {
	return s.location, s.getLocationErr
}

func (s *stateFactStoreStub) GetConnectionsFromLocation(_ context.Context, _ statedb.GetConnectionsFromLocationParams) ([]statedb.GetConnectionsFromLocationRow, error) {
	return s.connections, s.connectionsErr
}

func (s *stateFactStoreStub) GetPlayerCharacterByID(_ context.Context, _ pgtype.UUID) (statedb.PlayerCharacter, error) {
	return s.player, nil
}

func (s *stateFactStoreStub) UpdatePlayerLocation(_ context.Context, arg statedb.UpdatePlayerLocationParams) (statedb.PlayerCharacter, error) {
	s.lastUpdate = arg
	if s.updateErr != nil {
		return statedb.PlayerCharacter{}, s.updateErr
	}
	return statedb.PlayerCharacter{}, nil
}
func (s *stateFactStoreStub) SetLocationPlayerKnown(_ context.Context, _ pgtype.UUID) error {
	return s.knownErr
}
func (s *stateFactStoreStub) SetLocationPlayerVisited(_ context.Context, _ pgtype.UUID) error {
	return s.visitedErr
}

func TestReviseWorldFactScopesToCampaign(t *testing.T) {
	campaignID := uuid.New()
	otherCampaignID := uuid.New()
	factID := uuid.New()
	store := NewStateStore(&stateFactStoreStub{fact: statedb.WorldFact{ID: dbutil.ToPgtype(factID), CampaignID: dbutil.ToPgtype(otherCampaignID), Fact: "old", Category: "history"}})
	_, err := store.ReviseWorldFact(context.Background(), domain.ReviseWorldFactCommand{CampaignID: campaignID, FactID: factID, NewFact: "new"})
	if !errors.Is(err, domain.ErrWorldFactNotFound) {
		t.Fatalf("err = %v, want not found", err)
	}
}

func TestReviseWorldFactAlreadySuperseded(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	supersedingID := uuid.New()
	store := NewStateStore(&stateFactStoreStub{fact: statedb.WorldFact{ID: dbutil.ToPgtype(factID), CampaignID: dbutil.ToPgtype(campaignID), Fact: "old", Category: "history", SupersededBy: dbutil.ToPgtype(supersedingID)}})
	_, err := store.ReviseWorldFact(context.Background(), domain.ReviseWorldFactCommand{CampaignID: campaignID, FactID: factID, NewFact: "new"})
	if !errors.Is(err, domain.ErrWorldFactSuperseded) {
		t.Fatalf("err = %v, want superseded", err)
	}
}

func TestReviseWorldFactPropagatesKnownAndReveals(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	newFactID := uuid.New()
	stub := &stateFactStoreStub{
		fact:    statedb.WorldFact{ID: dbutil.ToPgtype(factID), CampaignID: dbutil.ToPgtype(campaignID), Fact: "old", Category: "history", Source: "established", PlayerKnown: true},
		newFact: statedb.WorldFact{ID: dbutil.ToPgtype(newFactID), CampaignID: dbutil.ToPgtype(campaignID), Fact: "new", Category: "history", Source: "established", PlayerKnown: true},
	}
	store := NewStateStore(stub)
	result, err := store.ReviseWorldFact(context.Background(), domain.ReviseWorldFactCommand{CampaignID: campaignID, FactID: factID, NewFact: "new", RevealToPlayer: true})
	if err != nil {
		t.Fatalf("ReviseWorldFact() error = %v", err)
	}
	if !result.PlayerKnownPropagated {
		t.Fatal("expected player-known propagation")
	}
	if result.NewFact.ID != newFactID {
		t.Fatalf("new fact id = %s, want %s", result.NewFact.ID, newFactID)
	}
	if !stub.lastSupersede.Reveal {
		t.Fatal("expected reveal flag in supersede params")
	}
}

func TestReviseWorldFactMapsNoRowsToSuperseded(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	store := NewStateStore(&stateFactStoreStub{fact: statedb.WorldFact{ID: dbutil.ToPgtype(factID), CampaignID: dbutil.ToPgtype(campaignID), Fact: "old", Category: "history"}, supersedeErr: pgx.ErrNoRows})
	_, err := store.ReviseWorldFact(context.Background(), domain.ReviseWorldFactCommand{CampaignID: campaignID, FactID: factID, NewFact: "new"})
	if !errors.Is(err, domain.ErrWorldFactSuperseded) {
		t.Fatalf("err = %v, want superseded", err)
	}
}

func TestRevealLocationScopesToCampaign(t *testing.T) {
	campaignID := uuid.New()
	locationID := uuid.New()
	store := NewStateStore(&stateFactStoreStub{getLocationErr: pgx.ErrNoRows})
	_, err := store.RevealLocation(context.Background(), RevealLocationCommand{CampaignID: campaignID, LocationID: locationID})
	if err == nil || err.Error() != "location not found in campaign" {
		t.Fatalf("err = %v", err)
	}
}

func TestMovePlayerBestEffortWarnings(t *testing.T) {
	campaignID := uuid.New()
	currentID := uuid.New()
	targetID := uuid.New()
	playerID := uuid.New()
	stub := &stateFactStoreStub{
		location:    statedb.Location{ID: dbutil.ToPgtype(targetID), CampaignID: dbutil.ToPgtype(campaignID), Name: "Harbor", Description: pgtype.Text{String: "Busy docks", Valid: true}},
		player:      statedb.PlayerCharacter{ID: dbutil.ToPgtype(playerID), CampaignID: dbutil.ToPgtype(campaignID)},
		connections: []statedb.GetConnectionsFromLocationRow{{ConnectedLocationID: dbutil.ToPgtype(targetID), TravelTime: pgtype.Text{String: "30 minutes", Valid: true}}},
		visitedErr:  errors.New("visited failed"),
	}
	store := NewStateStore(stub)
	res, err := store.MovePlayer(context.Background(), MovePlayerCommand{CampaignID: campaignID, PlayerCharacterID: playerID, CurrentLocationID: currentID, TargetLocationID: targetID})
	if err != nil {
		t.Fatalf("MovePlayer() err = %v", err)
	}
	if res.VisitedMarked {
		t.Fatal("expected visited warning")
	}
	if res.VisitedWarning == "" {
		t.Fatal("expected visited warning text")
	}
}
