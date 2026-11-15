package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

type startupScriptedResponse struct {
	resp *llm.Response
	err  error
}

type startupScriptedProvider struct {
	t         *testing.T
	scripts   []startupScriptedResponse
	calls     int
	lastTools []llm.Tool
}

func newStartupScriptedProvider(t *testing.T, scripts ...startupScriptedResponse) *startupScriptedProvider {
	t.Helper()
	return &startupScriptedProvider{t: t, scripts: scripts}
}

func (p *startupScriptedProvider) Complete(_ context.Context, _ []llm.Message, offered []llm.Tool) (*llm.Response, error) {
	p.t.Helper()
	if p.calls >= len(p.scripts) {
		p.t.Fatalf("Complete called %d time(s), but only %d response(s) were configured", p.calls+1, len(p.scripts))
	}
	p.lastTools = offered
	script := p.scripts[p.calls]
	p.calls++
	return script.resp, script.err
}

func (p *startupScriptedProvider) Stream(_ context.Context, _ []llm.Message, _ []llm.Tool) (<-chan llm.StreamChunk, error) {
	p.t.Helper()
	p.t.Fatal("unexpected Stream call in startup handler test")
	return nil, nil
}

type startupStubQuerier struct {
	statedb.Querier

	createCampaignFn          func(context.Context, statedb.CreateCampaignParams) (statedb.Campaign, error)
	createFactionFn           func(context.Context, statedb.CreateFactionParams) (statedb.Faction, error)
	createLocationFn          func(context.Context, statedb.CreateLocationParams) (statedb.Location, error)
	createNPCFn               func(context.Context, statedb.CreateNPCParams) (statedb.Npc, error)
	createFactFn              func(context.Context, statedb.CreateFactParams) (statedb.WorldFact, error)
	createItemFn              func(context.Context, statedb.CreateItemParams) (statedb.Item, error)
	createRelationshipFn      func(context.Context, statedb.CreateRelationshipParams) (statedb.EntityRelationship, error)
	setRelationshipAwareFn    func(context.Context, pgtype.UUID) error
	createPlayerCharacterFn   func(context.Context, statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error)
	createSessionLogFn        func(context.Context, statedb.CreateSessionLogParams) (statedb.SessionLog, error)
	listLocationsByCampaignFn func(context.Context, pgtype.UUID) ([]statedb.Location, error)

	createItemCalls              int
	createRelationshipCalls      int
	setRelationshipAwareCalls    int
	lastCreateItemParams         statedb.CreateItemParams
	lastCreateRelationshipParams statedb.CreateRelationshipParams
	lastSetRelationshipAwareID   pgtype.UUID
}

func (s *startupStubQuerier) CreateCampaign(ctx context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
	if s.createCampaignFn != nil {
		return s.createCampaignFn(ctx, arg)
	}
	return statedb.Campaign{}, nil
}

func (s *startupStubQuerier) CreateFaction(ctx context.Context, arg statedb.CreateFactionParams) (statedb.Faction, error) {
	if s.createFactionFn != nil {
		return s.createFactionFn(ctx, arg)
	}
	return statedb.Faction{}, nil
}

func (s *startupStubQuerier) CreateLocation(ctx context.Context, arg statedb.CreateLocationParams) (statedb.Location, error) {
	if s.createLocationFn != nil {
		return s.createLocationFn(ctx, arg)
	}
	return statedb.Location{}, nil
}

func (s *startupStubQuerier) CreateNPC(ctx context.Context, arg statedb.CreateNPCParams) (statedb.Npc, error) {
	if s.createNPCFn != nil {
		return s.createNPCFn(ctx, arg)
	}
	return statedb.Npc{}, nil
}

func (s *startupStubQuerier) CreateFact(ctx context.Context, arg statedb.CreateFactParams) (statedb.WorldFact, error) {
	if s.createFactFn != nil {
		return s.createFactFn(ctx, arg)
	}
	return statedb.WorldFact{}, nil
}

func (s *startupStubQuerier) CreateItem(ctx context.Context, arg statedb.CreateItemParams) (statedb.Item, error) {
	s.createItemCalls++
	s.lastCreateItemParams = arg
	if s.createItemFn != nil {
		return s.createItemFn(ctx, arg)
	}
	return statedb.Item{}, nil
}

func (s *startupStubQuerier) CreateRelationship(ctx context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error) {
	s.createRelationshipCalls++
	s.lastCreateRelationshipParams = arg
	if s.createRelationshipFn != nil {
		return s.createRelationshipFn(ctx, arg)
	}
	return statedb.EntityRelationship{}, nil
}

func (s *startupStubQuerier) SetRelationshipPlayerAware(ctx context.Context, id pgtype.UUID) error {
	s.setRelationshipAwareCalls++
	s.lastSetRelationshipAwareID = id
	if s.setRelationshipAwareFn != nil {
		return s.setRelationshipAwareFn(ctx, id)
	}
	return nil
}

func (s *startupStubQuerier) CreatePlayerCharacter(ctx context.Context, arg statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error) {
	if s.createPlayerCharacterFn != nil {
		return s.createPlayerCharacterFn(ctx, arg)
	}
	return statedb.PlayerCharacter{}, nil
}

func (s *startupStubQuerier) CreateSessionLog(ctx context.Context, arg statedb.CreateSessionLogParams) (statedb.SessionLog, error) {
	if s.createSessionLogFn != nil {
		return s.createSessionLogFn(ctx, arg)
	}
	return statedb.SessionLog{}, nil
}

func (s *startupStubQuerier) ListLocationsByCampaign(ctx context.Context, campaignID pgtype.UUID) ([]statedb.Location, error) {
	if s.listLocationsByCampaignFn != nil {
		return s.listLocationsByCampaignFn(ctx, campaignID)
	}
	return nil, nil
}

func newStartupRouter(h *StartupHandlers, authenticated bool) *chi.Mux {
	r := chi.NewRouter()
	if authenticated {
		authMW := auth.NewNoOpMiddleware(uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		r.Use(authMW.Authenticate)
	}
	r.Route("/api/v1/campaigns/start", func(r chi.Router) {
		r.Post("/campaign-interview", h.StartCampaignInterview)
		r.Post("/campaign-interview/{sessionID}", h.StepCampaignInterview)
		r.Post("/proposals", h.GenerateCampaignProposals)
		r.Post("/name", h.GenerateCampaignName)
		r.Post("/character-interview", h.StartCharacterInterview)
		r.Post("/character-interview/{sessionID}", h.StepCharacterInterview)
		r.Post("/world", h.BuildWorld)
	})
	return r
}

func TestCampaignInterview_StartAndStep(t *testing.T) {
	provider := newStartupScriptedProvider(t,
		startupScriptedResponse{resp: &llm.Response{Content: "What kind of world do you want to explore?"}},
		startupScriptedResponse{resp: &llm.Response{
			Content: "Perfect. Here's your campaign profile.",
			ToolCalls: []llm.ToolCall{{
				ID:   "call_1",
				Name: "extract_campaign_profile",
				Arguments: map[string]any{
					"genre":                "dark fantasy",
					"tone":                 "grim and tense",
					"themes":               []any{"survival", "betrayal"},
					"world_type":           "war-torn kingdom",
					"danger_level":         "high",
					"political_complexity": "complex",
				},
			}},
		}},
	)
	h := NewStartupHandlers(provider, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/campaign-interview", bytes.NewBufferString(`{}`))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, body = %s", startRec.Code, startRec.Body.String())
	}

	startResp := decodeStartupJSON[api.CampaignInterviewResponse](t, startRec)
	if startResp.SessionID == "" {
		t.Fatal("expected session_id")
	}
	if startResp.Done {
		t.Fatal("start response unexpectedly done")
	}
	if startResp.Message != "What kind of world do you want to explore?" {
		t.Fatalf("start message = %q", startResp.Message)
	}

	stepBody := mustStartupJSON(t, api.InterviewStepRequest{Input: "Dark fantasy with court intrigue."})
	stepReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/campaign-interview/"+startResp.SessionID, bytes.NewReader(stepBody))
	stepReq.Header.Set("Content-Type", "application/json")
	stepRec := httptest.NewRecorder()
	router.ServeHTTP(stepRec, stepReq)
	if stepRec.Code != http.StatusOK {
		t.Fatalf("step status = %d, body = %s", stepRec.Code, stepRec.Body.String())
	}

	stepResp := decodeStartupJSON[api.CampaignInterviewResponse](t, stepRec)
	if !stepResp.Done {
		t.Fatal("expected completed interview")
	}
	if stepResp.Profile == nil {
		t.Fatal("expected profile in completed response")
	}
	if stepResp.Profile.WorldType != "war-torn kingdom" {
		t.Fatalf("world_type = %q", stepResp.Profile.WorldType)
	}
}

func TestGenerateCampaignProposals_Success(t *testing.T) {
	provider := newStartupScriptedProvider(t, startupScriptedResponse{resp: &llm.Response{Content: `{"proposals":[{"name":"Ashfall","summary":"A frontier fortress holds the line against ancient horrors.","profile":{"genre":"dark fantasy","tone":"grim","themes":["survival","duty"],"world_type":"volcanic frontier","danger_level":"high","political_complexity":"moderate"}}]}`}})
	h := NewStartupHandlers(provider, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	body := mustStartupJSON(t, api.CampaignProposalsRequest{Genre: "fantasy", SettingStyle: "frontier", Tone: "grim"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/proposals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	resp := decodeStartupJSON[api.CampaignProposalsResponse](t, rec)
	if len(resp.Proposals) != 1 {
		t.Fatalf("proposal count = %d", len(resp.Proposals))
	}
	if resp.Proposals[0].Name != "Ashfall" {
		t.Fatalf("proposal name = %q", resp.Proposals[0].Name)
	}
	if resp.Proposals[0].Profile.DangerLevel != "high" {
		t.Fatalf("danger level = %q", resp.Proposals[0].Profile.DangerLevel)
	}
}

func TestGenerateCampaignName_Success(t *testing.T) {
	provider := newStartupScriptedProvider(t, startupScriptedResponse{resp: &llm.Response{Content: `{"name":"Ashfall"}`}})
	h := NewStartupHandlers(provider, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	body := mustStartupJSON(t, api.CampaignNameRequest{Profile: &api.CampaignProfile{Genre: "dark fantasy", Tone: "grim", Themes: []string{"survival"}, WorldType: "frontier", DangerLevel: "high", PoliticalComplexity: "moderate"}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/name", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	resp := decodeStartupJSON[api.CampaignNameResponse](t, rec)
	if resp.Name != "Ashfall" {
		t.Fatalf("name = %q", resp.Name)
	}
}

func TestCharacterInterview_StartAndStep(t *testing.T) {
	provider := newStartupScriptedProvider(t,
		startupScriptedResponse{resp: &llm.Response{Content: "Tell me about the hero you want to play."}},
		startupScriptedResponse{resp: &llm.Response{
			Content: "Excellent. Here's your hero.",
			ToolCalls: []llm.ToolCall{{
				ID:   "call_1",
				Name: "extract_character_profile",
				Arguments: map[string]any{
					"name":        "Kael",
					"concept":     "elven ranger",
					"background":  "Raised along the haunted frontier.",
					"personality": "quiet and vigilant",
					"motivations": []any{"Protect the innocent"},
					"strengths":   []any{"Tracking", "Archery"},
					"weaknesses":  []any{"Reckless when friends are threatened"},
				},
			}},
		}},
	)
	h := NewStartupHandlers(provider, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	startBody := mustStartupJSON(t, api.CharacterInterviewStartRequest{CampaignProfile: &api.CampaignProfile{Genre: "dark fantasy", Tone: "grim", Themes: []string{"survival"}, WorldType: "frontier", DangerLevel: "high", PoliticalComplexity: "moderate"}})
	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/character-interview", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, startReq)
	if startRec.Code != http.StatusOK {
		t.Fatalf("start status = %d, body = %s", startRec.Code, startRec.Body.String())
	}

	startResp := decodeStartupJSON[api.CharacterInterviewResponse](t, startRec)
	if startResp.SessionID == "" {
		t.Fatal("expected session_id")
	}
	if startResp.Done {
		t.Fatal("start response unexpectedly done")
	}

	stepBody := mustStartupJSON(t, api.InterviewStepRequest{Input: "An elven ranger named Kael."})
	stepReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/character-interview/"+startResp.SessionID, bytes.NewReader(stepBody))
	stepReq.Header.Set("Content-Type", "application/json")
	stepRec := httptest.NewRecorder()
	router.ServeHTTP(stepRec, stepReq)
	if stepRec.Code != http.StatusOK {
		t.Fatalf("step status = %d, body = %s", stepRec.Code, stepRec.Body.String())
	}

	stepResp := decodeStartupJSON[api.CharacterInterviewResponse](t, stepRec)
	if !stepResp.Done {
		t.Fatal("expected completed character interview")
	}
	if stepResp.Profile == nil || stepResp.Profile.Name != "Kael" {
		t.Fatalf("unexpected profile: %+v", stepResp.Profile)
	}
}

func TestBuildWorld_Success(t *testing.T) {
	provider := newStartupScriptedProvider(t,
		startupScriptedResponse{resp: &llm.Response{Content: `{"factions":[{"name":"Iron Guild","description":"Rules the mountain forges.","agenda":"Control trade","territory":"Northern peaks"}],"locations":[{"name":"Ironhold","description":"A fortress carved into obsidian cliffs.","region":"North","location_type":"city"}],"npcs":[{"name":"Marshal Vey","description":"Battle-scarred commander","personality":"stern","faction":"Iron Guild","location":"Ironhold"}],"world_facts":[{"fact":"Ash storms swallow the roads at dusk.","category":"environment"}],"starting_location":"Ironhold"}`}},
		startupScriptedResponse{resp: &llm.Response{Content: `{"narrative":"Ash falls across Ironhold as the gates groan open.","choices":["Enter the guild hall","Question Marshal Vey"]}`}},
	)
	campaignID := uuid.New()
	locationID := uuid.New()
	playerID := uuid.New()
	npcID := uuid.New()
	relationshipID := uuid.New()
	queries := &startupStubQuerier{
		createCampaignFn: func(_ context.Context, arg statedb.CreateCampaignParams) (statedb.Campaign, error) {
			now := time.Now()
			return statedb.Campaign{
				ID:          dbutil.ToPgtype(campaignID),
				Name:        arg.Name,
				Description: arg.Description,
				Genre:       arg.Genre,
				Tone:        arg.Tone,
				Themes:      arg.Themes,
				Status:      arg.Status,
				CreatedBy:   arg.CreatedBy,
				CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
			}, nil
		},
		createFactionFn: func(_ context.Context, _ statedb.CreateFactionParams) (statedb.Faction, error) {
			return statedb.Faction{ID: dbutil.ToPgtype(uuid.New())}, nil
		},
		createLocationFn: func(_ context.Context, arg statedb.CreateLocationParams) (statedb.Location, error) {
			id := locationID
			if arg.Name != "Ironhold" {
				id = uuid.New()
			}
			return statedb.Location{ID: dbutil.ToPgtype(id), Name: arg.Name}, nil
		},
		createNPCFn: func(_ context.Context, _ statedb.CreateNPCParams) (statedb.Npc, error) {
			return statedb.Npc{ID: dbutil.ToPgtype(uuid.New())}, nil
		},
		createFactFn: func(_ context.Context, _ statedb.CreateFactParams) (statedb.WorldFact, error) {
			return statedb.WorldFact{ID: dbutil.ToPgtype(uuid.New())}, nil
		},
		createPlayerCharacterFn: func(_ context.Context, _ statedb.CreatePlayerCharacterParams) (statedb.PlayerCharacter, error) {
			return statedb.PlayerCharacter{ID: dbutil.ToPgtype(playerID)}, nil
		},
		createItemFn: func(_ context.Context, arg statedb.CreateItemParams) (statedb.Item, error) {
			return statedb.Item{ID: dbutil.ToPgtype(uuid.New()), Name: arg.Name}, nil
		},
		createRelationshipFn: func(_ context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error) {
			return statedb.EntityRelationship{ID: dbutil.ToPgtype(relationshipID), RelationshipType: arg.RelationshipType}, nil
		},
		createSessionLogFn: func(_ context.Context, _ statedb.CreateSessionLogParams) (statedb.SessionLog, error) {
			return statedb.SessionLog{ID: dbutil.ToPgtype(uuid.New())}, nil
		},
		listLocationsByCampaignFn: func(_ context.Context, _ pgtype.UUID) ([]statedb.Location, error) {
			return []statedb.Location{{ID: dbutil.ToPgtype(locationID), Name: "Ironhold"}}, nil
		},
	}
	h := NewStartupHandlers(provider, queries, log.Default(), nil)
	router := newStartupRouter(h, true)

	body := mustStartupJSON(t, api.WorldBuildRequest{
		Name:    "Ashfall",
		Summary: "A fortress city resists a cursed frontier.",
		Profile: &api.CampaignProfile{
			Genre:               "dark fantasy",
			Tone:                "grim",
			Themes:              []string{"survival", "duty"},
			WorldType:           "volcanic frontier",
			DangerLevel:         "high",
			PoliticalComplexity: "moderate",
		},
		CharacterProfile: &api.CharacterProfile{
			Name:        "Kael",
			Concept:     "elven ranger",
			Background:  "Frontier scout",
			Personality: "quiet and vigilant",
			Motivations: []string{"Protect Ironhold"},
			Strengths:   []string{"Tracking"},
			Weaknesses:  []string{"Impulsive loyalty"},
		},
		SpawnPackage: &api.CharacterSpawnPackage{Items: []api.StarterItem{{Name: "Rusty Lantern", Description: "Still burns.", ItemType: "tool", Quantity: 1}}, KnownFacts: []api.StarterKnownFact{{Fact: "Kael knows the ash roads.", Category: "background"}}, Relationships: []api.StarterRelationship{{TargetEntityType: "npc", TargetEntityID: npcID.String(), RelationshipType: "knows", Description: "Old patrol contact"}}},
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/world", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	resp := decodeStartupJSON[api.WorldBuildResponse](t, rec)
	if resp.Campaign.Name != "Ashfall" {
		t.Fatalf("campaign name = %q", resp.Campaign.Name)
	}
	if resp.OpeningScene.Narrative == "" {
		t.Fatal("expected opening scene narrative")
	}
	if len(resp.OpeningScene.Choices) != 2 {
		t.Fatalf("choices count = %d", len(resp.OpeningScene.Choices))
	}
	if queries.createItemCalls != 1 {
		t.Fatalf("CreateItem calls = %d, want 1", queries.createItemCalls)
	}
	if got := queries.lastCreateItemParams.PlayerCharacterID; got != dbutil.ToPgtype(playerID) {
		t.Errorf("starter item player id = %#v, want %#v", got, dbutil.ToPgtype(playerID))
	}
	if got := queries.lastCreateItemParams.Name; got != "Rusty Lantern" {
		t.Errorf("starter item name = %q, want Rusty Lantern", got)
	}
	if queries.createRelationshipCalls != 1 {
		t.Fatalf("CreateRelationship calls = %d, want 1", queries.createRelationshipCalls)
	}
	if got := queries.lastCreateRelationshipParams.TargetEntityID; got != dbutil.ToPgtype(npcID) {
		t.Errorf("starter relationship target id = %#v, want %#v", got, dbutil.ToPgtype(npcID))
	}
	if queries.setRelationshipAwareCalls != 1 {
		t.Fatalf("SetRelationshipPlayerAware calls = %d, want 1", queries.setRelationshipAwareCalls)
	}
	if got := queries.lastSetRelationshipAwareID; got != dbutil.ToPgtype(relationshipID) {
		t.Errorf("revealed relationship id = %#v, want %#v", got, dbutil.ToPgtype(relationshipID))
	}
}

func TestBuildWorld_InvalidSpawnRelationshipTargetReturnsBadRequest(t *testing.T) {
	provider := newStartupScriptedProvider(t)
	h := NewStartupHandlers(provider, &startupStubQuerier{}, log.Default(), nil)
	router := newStartupRouter(h, true)
	body := mustStartupJSON(t, api.WorldBuildRequest{Name: "Ashfall", Summary: "A fortress city resists a cursed frontier.", Profile: &api.CampaignProfile{Genre: "dark fantasy", Tone: "grim", Themes: []string{"survival"}, WorldType: "frontier", DangerLevel: "high", PoliticalComplexity: "moderate"}, CharacterProfile: &api.CharacterProfile{Name: "Kael", Concept: "elven ranger", Background: "Frontier scout", Personality: "quiet and vigilant", Motivations: []string{"Protect Ironhold"}, Strengths: []string{"Tracking"}, Weaknesses: []string{"Impulsive loyalty"}}, SpawnPackage: &api.CharacterSpawnPackage{Relationships: []api.StarterRelationship{{TargetEntityType: "npc", TargetEntityID: "not-a-uuid", RelationshipType: "knows"}}}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/world", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCampaignInterview_UnknownSession(t *testing.T) {
	h := NewStartupHandlers(newStartupScriptedProvider(t), nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	body := mustStartupJSON(t, api.InterviewStepRequest{Input: "hello"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/campaign-interview/unknown", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestCampaignInterview_EmptyInput(t *testing.T) {
	provider := newStartupScriptedProvider(t, startupScriptedResponse{resp: &llm.Response{Content: "What kind of world do you want to explore?"}})
	h := NewStartupHandlers(provider, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	startReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/campaign-interview", bytes.NewBufferString(`{}`))
	startReq.Header.Set("Content-Type", "application/json")
	startRec := httptest.NewRecorder()
	router.ServeHTTP(startRec, startReq)
	startResp := decodeStartupJSON[api.CampaignInterviewResponse](t, startRec)

	body := mustStartupJSON(t, api.InterviewStepRequest{Input: "   "})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/campaign-interview/"+startResp.SessionID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestGenerateCampaignProposals_MissingProvider(t *testing.T) {
	h := NewStartupHandlers(nil, nil, log.Default(), nil)
	router := newStartupRouter(h, true)

	body := mustStartupJSON(t, api.CampaignProposalsRequest{Genre: "fantasy", SettingStyle: "frontier", Tone: "grim"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/proposals", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestBuildWorld_RequiresAuth(t *testing.T) {
	provider := newStartupScriptedProvider(t)
	h := NewStartupHandlers(provider, &startupStubQuerier{}, log.Default(), nil)
	router := newStartupRouter(h, false)

	body := mustStartupJSON(t, api.WorldBuildRequest{Profile: &api.CampaignProfile{}, CharacterProfile: &api.CharacterProfile{}})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/world", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func mustStartupJSON(t *testing.T, v any) []byte {
	t.Helper()
	body, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal request body: %v", err)
	}
	return body
}

func decodeStartupJSON[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.NewDecoder(rec.Body).Decode(&v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
	return v
}

func TestBuildWorld_RejectsIncompleteSupportedSRDCharacterCreation(t *testing.T) {
	provider := newStartupScriptedProvider(t)
	q := &startupStubQuerier{createCampaignFn: func(context.Context, statedb.CreateCampaignParams) (statedb.Campaign, error) {
		t.Fatal("versioned request reached legacy persistence")
		return statedb.Campaign{}, nil
	}}
	h := NewStartupHandlers(provider, q, log.Default(), nil)
	router := newStartupRouter(h, true)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/world", bytes.NewBufferString(`{"rules_version":"srd-5.2.1","profile":{},"character_profile":{}}`))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if provider.calls != 0 {
		t.Fatalf("legacy inference calls = %d", provider.calls)
	}
}

func (s *startupStubQuerier) RunInTransaction(ctx context.Context, fn func(statedb.Querier, statedb.DBTX) error) error {
	return fn(s, startupSetupDB{})
}

type startupSetupDB struct{ campaignRulesDB }

func (startupSetupDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}
func (startupSetupDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	return startupSetupRow{mode: args[0].(string)}
}

type startupSetupRow struct{ mode string }

func (r startupSetupRow) Scan(dest ...any) error { *dest[0].(*string) = r.mode; return nil }

func TestRulesCharacterInterviewSessionCorrectionsAndFailure(t *testing.T) {
	response := func(args map[string]any) *llm.Response {
		return &llm.Response{ToolCalls: []llm.ToolCall{{Name: "extract_character_profile", Arguments: args}}}
	}
	valid := map[string]any{"name": "Mira", "concept": "Human Cleric", "background": "Acolyte", "personality": "Patient", "motivations": []any{"Help"}, "strengths": []any{"Patience"}, "weaknesses": []any{"Pride"}, "rules_choices": map[string]any{"origin": "Human", "class": "Cleric", "background": "Acolyte"}}
	invalid := map[string]any{}
	for k, v := range valid {
		invalid[k] = v
	}
	delete(invalid, "rules_choices")
	invented := map[string]any{}
	for k, v := range valid {
		invented[k] = v
	}
	invented["abilities"] = []any{"Teleport"}
	provider := newStartupScriptedProvider(t, startupScriptedResponse{resp: &llm.Response{Content: "Choose your rules build."}}, startupScriptedResponse{resp: response(invalid)}, startupScriptedResponse{resp: response(invented)}, startupScriptedResponse{err: errors.New("unavailable")}, startupScriptedResponse{resp: response(valid)})
	h := NewStartupHandlers(provider, &startupStubQuerier{}, log.Default(), nil)
	router := newStartupRouter(h, true)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/character-interview", bytes.NewBufferString(`{"campaign_profile":{},"rules_version":"srd-5.2.1"}`)))
	if rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	start := decodeStartupJSON[api.CharacterInterviewResponse](t, rec)
	if _, ok := provider.lastTools[0].Parameters["properties"].(map[string]any)["rules_choices"]; !ok {
		t.Fatal("rules schema missing")
	}
	for i, want := range []int{200, 200, 500, 200} {
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/start/character-interview/"+start.SessionID, bytes.NewBufferString(`{"input":"Choose Human Cleric"}`)))
		if rec.Code != want {
			t.Fatalf("step %d: %d %s", i, rec.Code, rec.Body.String())
		}
		if want == 500 {
			continue
		}
		result := decodeStartupJSON[api.CharacterInterviewResponse](t, rec)
		if result.Done != (i == 3) {
			t.Fatalf("step %d: %+v", i, result)
		}
	}
}
