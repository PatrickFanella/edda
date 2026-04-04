package world

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/domain"
	"github.com/PatrickFanella/game-master/internal/llm"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

type stubQuerier struct {
	statedb.Querier

	createPlayerCharacterFn   func(context.Context, statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error)
	createCampaignFn          func(context.Context, statedb.CreateCampaignParams) (statedb.Campaign, error)
	createFactionFn           func(context.Context, statedb.CreateFactionParams) (statedb.Faction, error)
	createLocationFn          func(context.Context, statedb.CreateLocationParams) (statedb.Location, error)
	createNPCFn               func(context.Context, statedb.CreateNPCParams) (statedb.Npc, error)
	createFactFn              func(context.Context, statedb.CreateFactParams) (statedb.WorldFact, error)
	createSessionLogFn        func(context.Context, statedb.CreateSessionLogParams) (statedb.SessionLog, error)
	listLocationsByCampaignFn func(context.Context, pgtype.UUID) ([]statedb.Location, error)

	createPlayerCharacterCalls int
	createCampaignCalls        int
	createFactionCalls         int
	createLocationCalls        int
	createNPCCalls             int
	createFactCalls            int
	createSessionLogCalls      int
	listLocationsCalls         int

	lastCreatePlayerCharacterParams statedb.CreatePlayerCharacterParams
	lastCreateCampaignParams        statedb.CreateCampaignParams
	lastCreateFactionParams         statedb.CreateFactionParams
	lastCreateLocationParams        statedb.CreateLocationParams
	lastCreateNPCParams             statedb.CreateNPCParams
	lastCreateFactParams            statedb.CreateFactParams
	lastCreateSessionLogParams      statedb.CreateSessionLogParams
	lastListLocationsCampaignID     pgtype.UUID
}

func (s *stubQuerier) CreatePlayerCharacter(ctx context.Context, arg statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error) {
	s.createPlayerCharacterCalls++
	s.lastCreatePlayerCharacterParams = arg
	if s.createPlayerCharacterFn != nil {
		return s.createPlayerCharacterFn(ctx, arg)
	}
	return statedb.PlayerCharacter{}, nil
}

func (s *stubQuerier) CreateCampaign(ctx context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
	s.createCampaignCalls++
	s.lastCreateCampaignParams = arg
	if s.createCampaignFn != nil {
		return s.createCampaignFn(ctx, arg)
	}
	return statedb.Campaign{}, nil
}

func (s *stubQuerier) CreateFaction(ctx context.Context, arg statedb.CreateFactionParams) (statedb.Faction, error) {
	s.createFactionCalls++
	s.lastCreateFactionParams = arg
	if s.createFactionFn != nil {
		return s.createFactionFn(ctx, arg)
	}
	return statedb.Faction{}, nil
}

func (s *stubQuerier) CreateLocation(ctx context.Context, arg statedb.CreateLocationParams) (statedb.Location, error) {
	s.createLocationCalls++
	s.lastCreateLocationParams = arg
	if s.createLocationFn != nil {
		return s.createLocationFn(ctx, arg)
	}
	return statedb.Location{}, nil
}

func (s *stubQuerier) CreateNPC(ctx context.Context, arg statedb.CreateNPCParams) (statedb.Npc, error) {
	s.createNPCCalls++
	s.lastCreateNPCParams = arg
	if s.createNPCFn != nil {
		return s.createNPCFn(ctx, arg)
	}
	return statedb.Npc{}, nil
}

func (s *stubQuerier) CreateFact(ctx context.Context, arg statedb.CreateFactParams) (statedb.WorldFact, error) {
	s.createFactCalls++
	s.lastCreateFactParams = arg
	if s.createFactFn != nil {
		return s.createFactFn(ctx, arg)
	}
	return statedb.WorldFact{}, nil
}

func (s *stubQuerier) CreateSessionLog(ctx context.Context, arg statedb.CreateSessionLogParams) (statedb.SessionLog, error) {
	s.createSessionLogCalls++
	s.lastCreateSessionLogParams = arg
	if s.createSessionLogFn != nil {
		return s.createSessionLogFn(ctx, arg)
	}
	return statedb.SessionLog{}, nil
}

func (s *stubQuerier) ListLocationsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]statedb.Location, error) {
	s.listLocationsCalls++
	s.lastListLocationsCampaignID = campaignID
	if s.listLocationsByCampaignFn != nil {
		return s.listLocationsByCampaignFn(ctx, campaignID)
	}
	return nil, nil
}

type sequentialStubLLM struct {
	t       *testing.T
	scripts []stubSkeletonLLM
	calls   int
}

func (s *sequentialStubLLM) Complete(ctx context.Context, messages []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	s.t.Helper()
	if s.calls >= len(s.scripts) {
		s.t.Fatalf("Complete called %d time(s), but only %d response(s) were configured", s.calls+1, len(s.scripts))
	}
	resp, err := s.scripts[s.calls].Complete(ctx, messages, tools)
	s.calls++
	return resp, err
}

func (s *sequentialStubLLM) Stream(_ context.Context, _ []llm.Message, _ []llm.Tool) (<-chan llm.StreamChunk, error) {
	s.t.Helper()
	s.t.Fatal("unexpected Stream call in orchestrator test")
	return nil, nil
}

func pgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func testCharacterProfile() *CharacterProfile {
	return &CharacterProfile{
		Name:        "Kael",
		Concept:     "Elven ranger",
		Background:  "Raised on the frontier after a border war.",
		Personality: "Quiet and watchful",
		Motivations: []string{"Protect the wilds", "Find a missing mentor"},
		Strengths:   []string{"Tracking", "Archery"},
		Weaknesses:  []string{"Distrustful", "Reckless when family is threatened"},
	}
}

func TestPersistCharacterProfile_SuccessMapsFields(t *testing.T) {
	campaignID := uuid.New()
	userID := uuid.New()
	locationID := uuid.New()
	profile := testCharacterProfile()

	q := &stubQuerier{
		createPlayerCharacterFn: func(_ context.Context, arg statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error) {
			return statedb.PlayerCharacter{ID: pgUUID(uuid.New())}, nil
		},
	}

	_, err := PersistCharacterProfile(context.Background(), q, campaignID, userID, profile, &locationID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := q.lastCreatePlayerCharacterParams.Name; got != profile.Name {
		t.Errorf("name: got %q, want %q", got, profile.Name)
	}
	if got := q.lastCreatePlayerCharacterParams.Description; !got.Valid || !strings.Contains(got.String, profile.Concept) || !strings.Contains(got.String, profile.Background) {
		t.Errorf("description: got %#v, want concept/background included", got)
	}
	if got := q.lastCreatePlayerCharacterParams.Hp; got != 10 {
		t.Errorf("hp: got %d, want 10", got)
	}
	if got := q.lastCreatePlayerCharacterParams.MaxHp; got != 10 {
		t.Errorf("max hp: got %d, want 10", got)
	}
	if got := q.lastCreatePlayerCharacterParams.Level; got != 1 {
		t.Errorf("level: got %d, want 1", got)
	}
	if got := q.lastCreatePlayerCharacterParams.Experience; got != 0 {
		t.Errorf("xp: got %d, want 0", got)
	}
	if got := q.lastCreatePlayerCharacterParams.Status; got != "active" {
		t.Errorf("status: got %q, want active", got)
	}
	if got := q.lastCreatePlayerCharacterParams.CurrentLocationID; got != pgUUID(locationID) {
		t.Errorf("current location: got %#v, want %#v", got, pgUUID(locationID))
	}

	var stats map[string]any
	if err := json.Unmarshal(q.lastCreatePlayerCharacterParams.Stats, &stats); err != nil {
		t.Fatalf("unmarshal stats: %v", err)
	}
	if got := stats["concept"]; got != profile.Concept {
		t.Errorf("stats.concept: got %v, want %q", got, profile.Concept)
	}
	if got := stats["personality"]; got != profile.Personality {
		t.Errorf("stats.personality: got %v, want %q", got, profile.Personality)
	}
	gotMotivations, err := jsonArrayToStrings(stats["motivations"])
	if err != nil {
		t.Fatalf("stats.motivations: %v", err)
	}
	if !equalStrings(gotMotivations, profile.Motivations) {
		t.Errorf("stats.motivations: got %v, want %v", gotMotivations, profile.Motivations)
	}
	gotStrengths, err := jsonArrayToStrings(stats["strengths"])
	if err != nil {
		t.Fatalf("stats.strengths: %v", err)
	}
	if !equalStrings(gotStrengths, profile.Strengths) {
		t.Errorf("stats.strengths: got %v, want %v", gotStrengths, profile.Strengths)
	}
	gotWeaknesses, err := jsonArrayToStrings(stats["weaknesses"])
	if err != nil {
		t.Fatalf("stats.weaknesses: %v", err)
	}
	if !equalStrings(gotWeaknesses, profile.Weaknesses) {
		t.Errorf("stats.weaknesses: got %v, want %v", gotWeaknesses, profile.Weaknesses)
	}

	var abilities []string
	if err := json.Unmarshal(q.lastCreatePlayerCharacterParams.Abilities, &abilities); err != nil {
		t.Fatalf("unmarshal abilities: %v", err)
	}
	if !equalStrings(abilities, profile.Strengths) {
		t.Errorf("abilities: got %v, want %v", abilities, profile.Strengths)
	}
}

func TestPersistCharacterProfile_NilProfileReturnsError(t *testing.T) {
	_, err := PersistCharacterProfile(context.Background(), &stubQuerier{}, uuid.New(), uuid.New(), nil, nil)
	if err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestPersistCharacterProfile_NilStartingLocationLeavesInvalidCurrentLocation(t *testing.T) {
	q := &stubQuerier{}

	_, err := PersistCharacterProfile(context.Background(), q, uuid.New(), uuid.New(), testCharacterProfile(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := q.lastCreatePlayerCharacterParams.CurrentLocationID; got.Valid || got != (pgtype.UUID{}) {
		t.Errorf("current location: got %#v, want invalid zero pgtype.UUID", got)
	}
}

func TestSkeletonStoreAdapter_CreateFactionMapsFields(t *testing.T) {
	campaignID := uuid.New()
	factionID := uuid.New()
	q := &stubQuerier{
		createFactionFn: func(_ context.Context, arg statedb.CreateFactionParams) (statedb.Faction, error) {
			return statedb.Faction{ID: pgUUID(factionID)}, nil
		},
	}
	adapter := NewSkeletonStoreAdapter(q)

	gotID, err := adapter.CreateFaction(context.Background(), campaignID, SkeletonFaction{
		Name:        "Iron Guild",
		Description: "smiths",
		Agenda:      "control trade",
		Territory:   "mountains",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != factionID {
		t.Errorf("returned id: got %s, want %s", gotID, factionID)
	}

	params := q.lastCreateFactionParams
	if params.CampaignID != pgUUID(campaignID) {
		t.Errorf("campaign id: got %#v, want %#v", params.CampaignID, pgUUID(campaignID))
	}
	if params.Name != "Iron Guild" {
		t.Errorf("name: got %q, want Iron Guild", params.Name)
	}
	if params.Description != pgText("smiths") {
		t.Errorf("description: got %#v, want %#v", params.Description, pgText("smiths"))
	}
	if params.Agenda != pgText("control trade") {
		t.Errorf("agenda: got %#v, want %#v", params.Agenda, pgText("control trade"))
	}
	if params.Territory != pgText("mountains") {
		t.Errorf("territory: got %#v, want %#v", params.Territory, pgText("mountains"))
	}
}

func TestSkeletonStoreAdapter_CreateNPCNilIDsUseDefaults(t *testing.T) {
	campaignID := uuid.New()
	q := &stubQuerier{
		createNPCFn: func(_ context.Context, arg statedb.CreateNPCParams) (statedb.Npc, error) {
			return statedb.Npc{ID: pgUUID(uuid.New())}, nil
		},
	}
	adapter := NewSkeletonStoreAdapter(q)

	_, err := adapter.CreateNPC(context.Background(), campaignID, SkeletonNPC{
		Name:        "Mira",
		Description: "herbalist",
		Personality: "kind",
	}, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := q.lastCreateNPCParams
	if params.CampaignID != pgUUID(campaignID) {
		t.Errorf("campaign id: got %#v, want %#v", params.CampaignID, pgUUID(campaignID))
	}
	if params.FactionID.Valid {
		t.Errorf("faction id: got %#v, want invalid", params.FactionID)
	}
	if params.LocationID.Valid {
		t.Errorf("location id: got %#v, want invalid", params.LocationID)
	}
	if !params.Alive {
		t.Error("alive: got false, want true")
	}
	if params.Disposition != 50 {
		t.Errorf("disposition: got %d, want 50", params.Disposition)
	}
}

func TestSkeletonStoreAdapter_CreateWorldFactSetsSource(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	q := &stubQuerier{
		createFactFn: func(_ context.Context, arg statedb.CreateFactParams) (statedb.WorldFact, error) {
			return statedb.WorldFact{ID: pgUUID(factID)}, nil
		},
	}
	adapter := NewSkeletonStoreAdapter(q)

	gotID, err := adapter.CreateWorldFact(context.Background(), campaignID, SkeletonFact{Fact: "Dragons vanished", Category: "history"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != factID {
		t.Errorf("returned id: got %s, want %s", gotID, factID)
	}
	if got := q.lastCreateFactParams.Source; got != worldGenerationSource {
		t.Errorf("source: got %q, want %q", got, worldGenerationSource)
	}
}

func TestSkeletonStoreAdapter_SaveSessionLogMapsDomainFields(t *testing.T) {
	campaignID := uuid.New()
	locationID := uuid.New()
	npc1 := uuid.New()
	npc2 := uuid.New()
	toolCalls := json.RawMessage(`[{"name":"roll_dice"}]`)
	q := &stubQuerier{
		createSessionLogFn: func(_ context.Context, arg statedb.CreateSessionLogParams) (statedb.SessionLog, error) {
			return statedb.SessionLog{ID: pgUUID(uuid.New())}, nil
		},
	}
	adapter := NewSkeletonStoreAdapter(q)

	err := adapter.SaveSessionLog(context.Background(), domain.SessionLog{
		ID:           uuid.New(),
		CampaignID:   campaignID,
		TurnNumber:   7,
		PlayerInput:  "Scout the ruins",
		InputType:    domain.InputTypeNarrative,
		LLMResponse:  "You creep through the ruins.",
		ToolCalls:    toolCalls,
		LocationID:   &locationID,
		NPCsInvolved: []uuid.UUID{npc1, npc2},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	params := q.lastCreateSessionLogParams
	if params.CampaignID != pgUUID(campaignID) {
		t.Errorf("campaign id: got %#v, want %#v", params.CampaignID, pgUUID(campaignID))
	}
	if params.TurnNumber != 7 {
		t.Errorf("turn number: got %d, want 7", params.TurnNumber)
	}
	if params.PlayerInput != "Scout the ruins" {
		t.Errorf("player input: got %q, want %q", params.PlayerInput, "Scout the ruins")
	}
	if params.InputType != string(domain.InputTypeNarrative) {
		t.Errorf("input type: got %q, want %q", params.InputType, domain.InputTypeNarrative)
	}
	if params.LlmResponse != "You creep through the ruins." {
		t.Errorf("llm response: got %q, want %q", params.LlmResponse, "You creep through the ruins.")
	}
	if string(params.ToolCalls) != string(toolCalls) {
		t.Errorf("tool calls: got %s, want %s", string(params.ToolCalls), string(toolCalls))
	}
	if params.LocationID != pgUUID(locationID) {
		t.Errorf("location id: got %#v, want %#v", params.LocationID, pgUUID(locationID))
	}
	if len(params.NpcsInvolved) != 2 {
		t.Fatalf("npcs involved len: got %d, want 2", len(params.NpcsInvolved))
	}
	if params.NpcsInvolved[0] != pgUUID(npc1) {
		t.Errorf("npcs involved[0]: got %#v, want %#v", params.NpcsInvolved[0], pgUUID(npc1))
	}
	if params.NpcsInvolved[1] != pgUUID(npc2) {
		t.Errorf("npcs involved[1]: got %#v, want %#v", params.NpcsInvolved[1], pgUUID(npc2))
	}
}

func TestOrchestrator_RunSuccess(t *testing.T) {
	campaignID := uuid.New()
	userID := uuid.New()
	startingLocationID := uuid.New()
	characterID := uuid.New()
	profile := testProfile()
	character := testCharacterProfile()
	campaign := statedb.Campaign{
		ID:          pgUUID(campaignID),
		Name:        "Ashfall",
		Description: pgText("A frontier on the brink."),
		Status:      "active",
		CreatedBy:   pgUUID(userID),
	}

	q := &stubQuerier{
		createCampaignFn: func(_ context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
			return campaign, nil
		},
		createFactionFn: func(_ context.Context, arg statedb.CreateFactionParams) (statedb.Faction, error) {
			return statedb.Faction{ID: pgUUID(uuid.New())}, nil
		},
		createLocationFn: func(_ context.Context, arg statedb.CreateLocationParams) (statedb.Location, error) {
			return statedb.Location{ID: pgUUID(uuid.New())}, nil
		},
		createNPCFn: func(_ context.Context, arg statedb.CreateNPCParams) (statedb.Npc, error) {
			return statedb.Npc{ID: pgUUID(uuid.New())}, nil
		},
		createFactFn: func(_ context.Context, arg statedb.CreateFactParams) (statedb.WorldFact, error) {
			return statedb.WorldFact{ID: pgUUID(uuid.New())}, nil
		},
		listLocationsByCampaignFn: func(_ context.Context, gotCampaignID pgtype.UUID) ([]statedb.Location, error) {
			return []statedb.Location{{ID: pgUUID(startingLocationID), CampaignID: gotCampaignID, Name: validResponse().StartingLocation}}, nil
		},
		createPlayerCharacterFn: func(_ context.Context, arg statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error) {
			return statedb.PlayerCharacter{ID: pgUUID(characterID)}, nil
		},
		createSessionLogFn: func(_ context.Context, arg statedb.CreateSessionLogParams) (statedb.SessionLog, error) {
			return statedb.SessionLog{ID: pgUUID(uuid.New())}, nil
		},
	}

	provider := &sequentialStubLLM{
		t: t,
		scripts: []stubSkeletonLLM{
			{content: mustJSON(t, validResponse())},
			{content: mustSceneJSON(t, validSceneResponse())},
		},
	}

	orch := NewOrchestrator(provider, q)
	var stages []string
	result, err := orch.Run(context.Background(), OrchestratorInput{
		Name:             "Ashfall",
		Summary:          "A frontier on the brink.",
		Profile:          profile,
		CharacterProfile: character,
		UserID:           userID,
	}, func(stage string) {
		stages = append(stages, stage)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Campaign.ID != campaign.ID {
		t.Errorf("campaign id: got %#v, want %#v", result.Campaign.ID, campaign.ID)
	}
	if result.Scene == nil {
		t.Fatal("expected scene result")
	}
	if got := result.Scene.Narrative; got != validSceneResponse().Narrative {
		t.Errorf("scene narrative: got %q, want %q", got, validSceneResponse().Narrative)
	}
	if result.StartingLocationID != startingLocationID {
		t.Errorf("starting location id: got %s, want %s", result.StartingLocationID, startingLocationID)
	}

	expectedStages := []string{
		"Creating campaign…",
		"Forging the world…",
		"Bringing your character to life…",
		"Setting the scene…",
	}
	if !equalStrings(stages, expectedStages) {
		t.Errorf("stages: got %v, want %v", stages, expectedStages)
	}
	if q.createPlayerCharacterCalls != 1 {
		t.Errorf("create player character calls: got %d, want 1", q.createPlayerCharacterCalls)
	}
	if q.createSessionLogCalls != 1 {
		t.Errorf("create session log calls: got %d, want 1", q.createSessionLogCalls)
	}
	if q.lastListLocationsCampaignID != campaign.ID {
		t.Errorf("list locations campaign id: got %#v, want %#v", q.lastListLocationsCampaignID, campaign.ID)
	}
	if got := q.lastCreatePlayerCharacterParams.CurrentLocationID; got != pgUUID(startingLocationID) {
		t.Errorf("persisted current location: got %#v, want %#v", got, pgUUID(startingLocationID))
	}
}

func TestOrchestrator_RunNilCampaignProfileReturnsError(t *testing.T) {
	orch := NewOrchestrator(&stubSkeletonLLM{content: "{}"}, &stubQuerier{})

	_, err := orch.Run(context.Background(), OrchestratorInput{
		CharacterProfile: testCharacterProfile(),
		UserID:           uuid.New(),
	}, nil)
	if err == nil {
		t.Fatal("expected error for nil campaign profile")
	}
}

func TestOrchestrator_RunNilCharacterProfileReturnsError(t *testing.T) {
	orch := NewOrchestrator(&stubSkeletonLLM{content: "{}"}, &stubQuerier{})

	_, err := orch.Run(context.Background(), OrchestratorInput{
		Profile: testProfile(),
		UserID:  uuid.New(),
	}, nil)
	if err == nil {
		t.Fatal("expected error for nil character profile")
	}
}

func TestOrchestrator_RunCreateCampaignError(t *testing.T) {
	wantErr := errors.New("insert failed")
	q := &stubQuerier{
		createCampaignFn: func(_ context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
			return statedb.Campaign{}, wantErr
		},
	}
	orch := NewOrchestrator(&stubSkeletonLLM{content: "{}"}, q)

	_, err := orch.Run(context.Background(), OrchestratorInput{
		Name:             "Ashfall",
		Summary:          "A frontier on the brink.",
		Profile:          testProfile(),
		CharacterProfile: testCharacterProfile(),
		UserID:           uuid.New(),
	}, nil)
	if err == nil {
		t.Fatal("expected error from create campaign")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got %v, want wrapped %v", err, wantErr)
	}
	if !strings.Contains(err.Error(), "orchestrator: create campaign") {
		t.Fatalf("error: got %q, want create campaign context", err.Error())
	}
}

func jsonArrayToStrings(v any) ([]string, error) {
	items, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("got %T, want []any", v)
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("array item got %T, want string", item)
		}
		out = append(out, s)
	}
	return out, nil
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
