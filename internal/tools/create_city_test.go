package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

type stubCityStore struct {
	locationsByID         map[[16]byte]statedb.Location
	locationsByCampaign   []statedb.Location
	createLocationCalls   []statedb.CreateLocationParams
	updateLocationCalls   []statedb.UpdateLocationParams
	createLocationResults []statedb.Location
	updateLocationResult  statedb.Location
	getLocationErr        error
	listLocationsErr      error
	createLocationErr     error
	updateLocationErr     error
}

func (s *stubCityStore) CreateLocation(_ context.Context, arg statedb.CreateLocationParams) (statedb.Location, error) {
	if s.createLocationErr != nil {
		return statedb.Location{}, s.createLocationErr
	}
	s.createLocationCalls = append(s.createLocationCalls, arg)
	if len(s.createLocationResults) > 0 {
		out := s.createLocationResults[0]
		s.createLocationResults = s.createLocationResults[1:]
		return out, nil
	}
	return statedb.Location{
		ID:         dbutil.ToPgtype(uuid.New()),
		CampaignID: arg.CampaignID,
		Name:       arg.Name,
	}, nil
}

func (s *stubCityStore) UpdateLocation(_ context.Context, arg statedb.UpdateLocationParams) (statedb.Location, error) {
	if s.updateLocationErr != nil {
		return statedb.Location{}, s.updateLocationErr
	}
	s.updateLocationCalls = append(s.updateLocationCalls, arg)
	if s.updateLocationResult.ID.Valid {
		return s.updateLocationResult, nil
	}
	return statedb.Location{
		ID:         arg.ID,
		CampaignID: dbutil.ToPgtype(uuid.New()),
		Name:       arg.Name,
	}, nil
}

func (s *stubCityStore) GetLocationByID(_ context.Context, id pgtype.UUID) (statedb.Location, error) {
	if s.getLocationErr != nil {
		return statedb.Location{}, s.getLocationErr
	}
	location, ok := s.locationsByID[id.Bytes]
	if !ok {
		return statedb.Location{}, errors.New("location not found")
	}
	return location, nil
}

func (s *stubCityStore) ListLocationsByCampaign(_ context.Context, _ pgtype.UUID) ([]statedb.Location, error) {
	if s.listLocationsErr != nil {
		return nil, s.listLocationsErr
	}
	out := make([]statedb.Location, len(s.locationsByCampaign))
	copy(out, s.locationsByCampaign)
	return out, nil
}

func TestRegisterCreateCity(t *testing.T) {
	reg := NewRegistry()
	cityStore := &stubCityStore{}
	memStore := &stubMemoryStore{}
	embedder := &stubEmbedder{vector: []float32{0.1}}

	if err := RegisterCreateCity(reg, cityStore, memStore, embedder); err != nil {
		t.Fatalf("register create_city: %v", err)
	}

	tools := reg.List()
	if len(tools) != 1 {
		t.Fatalf("registered tool count = %d, want 1", len(tools))
	}
	if tools[0].Name != createCityToolName {
		t.Fatalf("tool name = %q, want %q", tools[0].Name, createCityToolName)
	}

	required, ok := tools[0].Parameters["required"].([]string)
	if !ok {
		t.Fatalf("required schema has unexpected type %T", tools[0].Parameters["required"])
	}
	requiredSet := make(map[string]struct{}, len(required))
	for _, field := range required {
		requiredSet[field] = struct{}{}
	}
	for _, field := range []string{"name", "description", "population", "districts", "landmarks", "governance", "economy_summary", "demographics", "location_id"} {
		if _, ok := requiredSet[field]; !ok {
			t.Fatalf("schema missing required field %q", field)
		}
	}
}

func TestCreateCityHandleCreateAndDistrictsSuccess(t *testing.T) {
	campaignID := uuid.New()
	parentLocationID := uuid.New()
	cityID := uuid.New()
	districtID := uuid.New()

	parentLocation := statedb.Location{
		ID:         dbutil.ToPgtype(parentLocationID),
		CampaignID: dbutil.ToPgtype(campaignID),
		Name:       "Northern Realm",
		Region:     pgtype.Text{String: "Northlands", Valid: true},
	}
	store := &stubCityStore{
		locationsByID: map[[16]byte]statedb.Location{
			dbutil.ToPgtype(parentLocationID).Bytes: parentLocation,
		},
		locationsByCampaign: []statedb.Location{
			parentLocation,
		},
		createLocationResults: []statedb.Location{
			{
				ID:         dbutil.ToPgtype(cityID),
				CampaignID: dbutil.ToPgtype(campaignID),
				Name:       "Ironhold",
			},
			{
				ID:         dbutil.ToPgtype(districtID),
				CampaignID: dbutil.ToPgtype(campaignID),
				Name:       "Foundry Ward",
			},
		},
	}
	memStore := &stubMemoryStore{}
	embedder := &stubEmbedder{vector: []float32{0.2, 0.3}}
	h := NewCreateCityHandler(store, memStore, embedder)

	got, err := h.Handle(context.Background(), map[string]any{
		"name":            "Ironhold",
		"description":     "A fortified industrial city.",
		"population":      120000,
		"districts":       []any{"Foundry Ward"},
		"landmarks":       []any{"Sky Bastion"},
		"governance":      "Council of Forgemasters",
		"economy_summary": "Steel and clockwork exports",
		"demographics": map[string]any{
			"dwarves": "40%",
		},
		"location_id":               parentLocationID.String(),
		"create_district_locations": true,
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(store.createLocationCalls) != 2 {
		t.Fatalf("CreateLocation call count = %d, want 2", len(store.createLocationCalls))
	}
	if store.createLocationCalls[0].LocationType.String != "city" {
		t.Fatalf("city location_type = %q, want city", store.createLocationCalls[0].LocationType.String)
	}
	if store.createLocationCalls[1].LocationType.String != "district" {
		t.Fatalf("district location_type = %q, want district", store.createLocationCalls[1].LocationType.String)
	}

	var cityProps map[string]any
	if err := json.Unmarshal(store.createLocationCalls[0].Properties, &cityProps); err != nil {
		t.Fatalf("unmarshal city properties: %v", err)
	}
	if gotDistricts, ok := cityProps["districts"].([]any); !ok || len(gotDistricts) != 1 || gotDistricts[0] != "Foundry Ward" {
		t.Fatalf("city properties districts = %#v, want [Foundry Ward]", cityProps["districts"])
	}
	if gotLandmarks, ok := cityProps["landmarks"].([]any); !ok || len(gotLandmarks) != 1 || gotLandmarks[0] != "Sky Bastion" {
		t.Fatalf("city properties landmarks = %#v, want [Sky Bastion]", cityProps["landmarks"])
	}
	if cityProps["governance"] != "Council of Forgemasters" {
		t.Fatalf("city properties governance = %#v, want Council of Forgemasters", cityProps["governance"])
	}
	if cityProps["economy_summary"] != "Steel and clockwork exports" {
		t.Fatalf("city properties economy_summary = %#v, want Steel and clockwork exports", cityProps["economy_summary"])
	}

	if got.Data["id"] != cityID.String() {
		t.Fatalf("result id = %v, want %s", got.Data["id"], cityID.String())
	}
	if got.Data["action"] != "created" {
		t.Fatalf("result action = %v, want created", got.Data["action"])
	}
	if memStore.lastParams.MemoryType != string(domain.MemoryTypeWorldFact) {
		t.Fatalf("CreateMemory memory_type = %q, want %q", memStore.lastParams.MemoryType, domain.MemoryTypeWorldFact)
	}
	if embedder.lastInput == "" {
		t.Fatal("expected embedder input to be populated")
	}
}

func TestCreateCityHandleUpdateExistingCity(t *testing.T) {
	campaignID := uuid.New()
	parentLocationID := uuid.New()
	cityID := uuid.New()

	existingCityProperties, err := json.Marshal(map[string]any{
		"parent_location_id": parentLocationID.String(),
	})
	if err != nil {
		t.Fatalf("marshal existing city props: %v", err)
	}

	parentLocation := statedb.Location{
		ID:         dbutil.ToPgtype(parentLocationID),
		CampaignID: dbutil.ToPgtype(campaignID),
		Name:       "Eastern March",
	}
	store := &stubCityStore{
		locationsByID: map[[16]byte]statedb.Location{
			dbutil.ToPgtype(parentLocationID).Bytes: parentLocation,
		},
		locationsByCampaign: []statedb.Location{
			parentLocation,
			{
				ID:           dbutil.ToPgtype(cityID),
				CampaignID:   dbutil.ToPgtype(campaignID),
				Name:         "Ironhold",
				LocationType: pgtype.Text{String: "city", Valid: true},
				Properties:   existingCityProperties,
			},
		},
		updateLocationResult: statedb.Location{
			ID:         dbutil.ToPgtype(cityID),
			CampaignID: dbutil.ToPgtype(campaignID),
			Name:       "Ironhold",
		},
	}

	h := NewCreateCityHandler(store, nil, nil)
	got, err := h.Handle(context.Background(), map[string]any{
		"name":            "Ironhold",
		"description":     "Updated description",
		"population":      130000,
		"districts":       []any{"Foundry Ward"},
		"landmarks":       []any{"Sky Bastion"},
		"governance":      "High Forge Council",
		"economy_summary": "Metalworks and banking",
		"demographics": map[string]any{
			"humans": "35%",
		},
		"location_id": parentLocationID.String(),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}

	if len(store.updateLocationCalls) != 1 {
		t.Fatalf("UpdateLocation call count = %d, want 1", len(store.updateLocationCalls))
	}
	if len(store.createLocationCalls) != 0 {
		t.Fatalf("CreateLocation call count = %d, want 0 for update path", len(store.createLocationCalls))
	}
	if got.Data["action"] != "updated" {
		t.Fatalf("result action = %v, want updated", got.Data["action"])
	}
}

func TestCreateCityValidationAndErrors(t *testing.T) {
	parentLocationID := uuid.New()
	baseArgs := map[string]any{
		"name":            "City",
		"description":     "desc",
		"population":      1000,
		"districts":       []any{"Old Quarter"},
		"landmarks":       []any{"Clocktower"},
		"governance":      "Mayor",
		"economy_summary": "Trade",
		"demographics": map[string]any{
			"humans": "60%",
		},
		"location_id": parentLocationID.String(),
	}

	t.Run("missing required field", func(t *testing.T) {
		h := NewCreateCityHandler(&stubCityStore{}, &stubMemoryStore{}, &stubEmbedder{vector: []float32{0.1}})
		args := copyArgs(baseArgs)
		delete(args, "districts")
		_, err := h.Handle(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "districts is required") {
			t.Fatalf("error = %v, want districts-required message", err)
		}
	})

	t.Run("invalid population type", func(t *testing.T) {
		h := NewCreateCityHandler(&stubCityStore{}, &stubMemoryStore{}, &stubEmbedder{vector: []float32{0.1}})
		args := copyArgs(baseArgs)
		args["population"] = "many"
		_, err := h.Handle(context.Background(), args)
		if err == nil || !strings.Contains(err.Error(), "population must be an integer") {
			t.Fatalf("error = %v, want population integer message", err)
		}
	})

	t.Run("invalid location_id reference", func(t *testing.T) {
		h := NewCreateCityHandler(
			&stubCityStore{
				locationsByID: map[[16]byte]statedb.Location{},
			},
			&stubMemoryStore{},
			&stubEmbedder{vector: []float32{0.1}},
		)
		_, err := h.Handle(context.Background(), copyArgs(baseArgs))
		if err == nil || !strings.Contains(err.Error(), "validate location_id") {
			t.Fatalf("error = %v, want validate location_id context", err)
		}
	})
}

var _ CityStore = (*stubCityStore)(nil)
