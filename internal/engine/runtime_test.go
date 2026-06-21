package engine

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/assembly"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/memory"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"github.com/google/uuid"
)

func TestMarshalAppliedToolCallsNilEncodesArray(t *testing.T) {
	got, err := marshalAppliedToolCalls(nil)
	if err != nil {
		t.Fatalf("marshalAppliedToolCalls(nil) error = %v", err)
	}
	if string(got) != "[]" {
		t.Fatalf("expected empty JSON array, got %s", got)
	}
}

func TestMarshalAppliedToolCallsPreservesEntries(t *testing.T) {
	got, err := marshalAppliedToolCalls([]AppliedToolCall{{
		Tool:      "skill_check",
		Arguments: json.RawMessage(`{"skill":"stealth"}`),
		Result:    json.RawMessage(`{"success":true}`),
	}})
	if err != nil {
		t.Fatalf("marshalAppliedToolCalls() error = %v", err)
	}
	if string(got) == "[]" {
		t.Fatal("expected non-empty marshaled tool calls")
	}
}

func TestFinalLocationIDFromApplied_DefaultLocationRemains(t *testing.T) {
	locationID := uuid.New()
	if got := finalLocationIDFromApplied(&locationID, nil); got == nil || *got != locationID {
		t.Fatalf("finalLocationIDFromApplied(nil) = %v, want %v", got, locationID)
	}
}

func TestFinalLocationIDFromApplied_MovePlayerOverridesDefault(t *testing.T) {
	defaultLocationID := uuid.New()
	newLocationID := uuid.New()
	got := finalLocationIDFromApplied(&defaultLocationID, []AppliedToolCall{{
		Tool:   "move_player",
		Result: json.RawMessage(`{"location_id":"` + newLocationID.String() + `"}`),
	}})
	if got == nil || *got != newLocationID {
		t.Fatalf("finalLocationIDFromApplied(move_player) = %v, want %v", got, newLocationID)
	}
}

func TestFinalLocationIDFromApplied_CreateLocationMoveHereOverridesDefault(t *testing.T) {
	defaultLocationID := uuid.New()
	newLocationID := uuid.New()
	got := finalLocationIDFromApplied(&defaultLocationID, []AppliedToolCall{{
		Tool:   "create_location",
		Result: json.RawMessage(`{"move_player_here":true,"location_id":"` + newLocationID.String() + `"}`),
	}})
	if got == nil || *got != newLocationID {
		t.Fatalf("finalLocationIDFromApplied(create_location move) = %v, want %v", got, newLocationID)
	}
}

func TestFinalLocationIDFromApplied_CreateLocationWithoutMoveKeepsDefault(t *testing.T) {
	defaultLocationID := uuid.New()
	newLocationID := uuid.New()
	got := finalLocationIDFromApplied(&defaultLocationID, []AppliedToolCall{{
		Tool:   "create_location",
		Result: json.RawMessage(`{"move_player_here":false,"location_id":"` + newLocationID.String() + `"}`),
	}})
	if got == nil || *got != defaultLocationID {
		t.Fatalf("finalLocationIDFromApplied(create_location no move) = %v, want %v", got, defaultLocationID)
	}
}

type testProvider struct{}

func (p *testProvider) Complete(_ context.Context, _ []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	var foundUpdateNPC bool
	var foundUpdatePlayerStats bool
	var foundAddExperience bool
	var foundLevelUp bool
	var foundAddAbility bool
	var foundRemoveAbility bool
	var foundUpdateQuest bool
	for _, tool := range tools {
		if tool.Name == "update_npc" {
			foundUpdateNPC = true
		}
		if tool.Name == "update_player_stats" {
			foundUpdatePlayerStats = true
		}
		if tool.Name == "add_experience" {
			foundAddExperience = true
		}
		if tool.Name == "level_up" {
			foundLevelUp = true
		}
		if tool.Name == "add_ability" {
			foundAddAbility = true
		}
		if tool.Name == "remove_ability" {
			foundRemoveAbility = true
		}
		if tool.Name == "update_quest" {
			foundUpdateQuest = true
		}
	}
	if !foundUpdateNPC {
		return nil, errors.New("update_npc tool not registered")
	}
	if !foundUpdatePlayerStats {
		return nil, errors.New("update_player_stats tool not registered")
	}
	if !foundAddExperience {
		return nil, errors.New("add_experience tool not registered")
	}
	if !foundLevelUp {
		return nil, errors.New("level_up tool not registered")
	}
	if !foundAddAbility {
		return nil, errors.New("add_ability tool not registered")
	}
	if !foundRemoveAbility {
		return nil, errors.New("remove_ability tool not registered")
	}
	if !foundUpdateQuest {
		return nil, errors.New("update_quest tool not registered")
	}
	return &llm.Response{
		Content: "",
		ToolCalls: []llm.ToolCall{
			{
				ID:   "1",
				Name: "update_npc",
				Arguments: map[string]any{
					"npc_id":      "00000000-0000-0000-0000-000000000001",
					"description": "updated via runtime registration test",
				},
			},
		},
	}, nil
}

func (p *testProvider) Stream(_ context.Context, _ []llm.Message, _ []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

type stubStateManager struct{}

func (s *stubStateManager) GetOrCreateDefaultUser(context.Context) (*domain.User, error) {
	return nil, nil
}
func (s *stubStateManager) CreateCampaign(context.Context, game.CreateCampaignParams) (*domain.Campaign, error) {
	return nil, nil
}
func (s *stubStateManager) GatherState(context.Context, uuid.UUID) (*game.GameState, error) {
	return &game.GameState{}, nil
}
func (s *stubStateManager) SaveSessionLog(context.Context, domain.SessionLog) error { return nil }
func (s *stubStateManager) ListRecentSessionLogs(context.Context, uuid.UUID, int) ([]domain.SessionLog, error) {
	return nil, nil
}
func (s *stubStateManager) GetCampaignByID(context.Context, uuid.UUID) (*domain.Campaign, error) {
	return nil, nil
}

type echoStatusProvider struct{}

func (p *echoStatusProvider) Complete(_ context.Context, messages []llm.Message, _ []llm.Tool) (*llm.Response, error) {
	return &llm.Response{Content: "done"}, nil
}

func (p *echoStatusProvider) Stream(_ context.Context, _ []llm.Message, _ []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

type contextAwareProvider struct{}

func (p *contextAwareProvider) Complete(_ context.Context, _ []llm.Message, tools []llm.Tool) (*llm.Response, error) {
	for _, tool := range tools {
		if tool.Name == "update_player_hp" {
			return &llm.Response{Content: "", ToolCalls: []llm.ToolCall{{ID: "1", Name: "update_player_hp", Arguments: map[string]any{"hp": 7}}}}, nil
		}
	}
	return &llm.Response{Content: "done"}, nil
}

func (p *contextAwareProvider) Stream(_ context.Context, _ []llm.Message, _ []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("not implemented")
}

type contextAwareStateManager struct {
	playerID   uuid.UUID
	locationID uuid.UUID
}

func (s *contextAwareStateManager) GetOrCreateDefaultUser(context.Context) (*domain.User, error) {
	return nil, nil
}
func (s *contextAwareStateManager) CreateCampaign(context.Context, game.CreateCampaignParams) (*domain.Campaign, error) {
	return nil, nil
}
func (s *contextAwareStateManager) GatherState(context.Context, uuid.UUID) (*game.GameState, error) {
	return &game.GameState{Player: domain.PlayerCharacter{ID: s.playerID, CurrentLocationID: &s.locationID}}, nil
}
func (s *contextAwareStateManager) SaveSessionLog(context.Context, domain.SessionLog) error {
	return nil
}
func (s *contextAwareStateManager) ListRecentSessionLogs(context.Context, uuid.UUID, int) ([]domain.SessionLog, error) {
	return nil, nil
}
func (s *contextAwareStateManager) GetCampaignByID(context.Context, uuid.UUID) (*domain.Campaign, error) {
	return nil, nil
}

type contextAwareCombatStore struct {
	mu           sync.Mutex
	playerID     uuid.UUID
	locationID   uuid.UUID
	ctxSeen      bool
	campaignID   uuid.UUID
	playerSeen   uuid.UUID
	locationSeen uuid.UUID
}

func (s *contextAwareCombatStore) GetPlayerCharacterByID(ctx context.Context, playerCharacterID uuid.UUID) (*domain.PlayerCharacter, error) {
	return &domain.PlayerCharacter{ID: playerCharacterID, HP: 3, MaxHP: 10}, nil
}
func (s *contextAwareCombatStore) UpdatePlayerCurrentHP(ctx context.Context, playerCharacterID uuid.UUID, hp int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ctxSeen = true
	if campaignID, ok := tools.CurrentCampaignIDFromContext(ctx); ok {
		s.campaignID = campaignID
	}
	if playerSeen, ok := tools.CurrentPlayerCharacterIDFromContext(ctx); ok {
		s.playerSeen = playerSeen
	}
	if locationSeen, ok := tools.CurrentLocationIDFromContext(ctx); ok {
		s.locationSeen = locationSeen
	}
	return nil
}

func TestProcessTurnStream_StreamToolReceivesEnrichedContext(t *testing.T) {
	playerID := uuid.New()
	locationID := uuid.New()
	campaignID := uuid.New()
	store := &contextAwareCombatStore{playerID: playerID, locationID: locationID}
	reg := tools.NewRegistry()
	if err := tools.RegisterUpdatePlayerHP(reg, store); err != nil {
		t.Fatalf("RegisterUpdatePlayerHP() error = %v", err)
	}
	e := &Engine{
		logger:    slog.Default(),
		state:     &contextAwareStateManager{playerID: playerID, locationID: locationID},
		assembler: assembly.NewContextAssembler(reg),
		processor: NewTurnProcessor(&contextAwareProvider{}, reg, tools.NewValidator(reg), nil),
	}

	ch, err := e.ProcessTurnStream(context.Background(), campaignID, "rest and recover")
	if err != nil {
		t.Fatalf("ProcessTurnStream() error = %v", err)
	}
	for ev := range ch {
		if ev.Type == "error" {
			t.Fatalf("unexpected stream error: %v", ev.Err)
		}
	}

	if !store.ctxSeen {
		t.Fatal("expected streamed tool execution to receive context")
	}
	if store.campaignID != campaignID {
		t.Fatalf("campaign id seen by tool = %v, want %v", store.campaignID, campaignID)
	}
	if store.playerSeen != playerID {
		t.Fatalf("player id seen by tool = %v, want %v", store.playerSeen, playerID)
	}
	if store.locationSeen != locationID {
		t.Fatalf("location id seen by tool = %v, want %v", store.locationSeen, locationID)
	}
}

func TestNewRegistersExpectedTools(t *testing.T) {
	e, err := New(nil, &testProvider{}, config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{ContextTokenBudget: 8000}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	tools := e.assembler.Tools()
	want := []string{
		"update_npc",
		"update_player_stats",
		"add_experience",
		"level_up",
		"add_ability",
		"remove_ability",
		"update_quest",
	}
	registered := make(map[string]bool, len(tools))
	for _, tool := range tools {
		registered[tool.Name] = true
	}
	for _, name := range want {
		if !registered[name] {
			t.Errorf("expected tool %q to be registered", name)
		}
	}
}

func TestProcessTurnStream_Error(t *testing.T) {
	e, err := New(nil, &testProvider{}, config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{ContextTokenBudget: 8000}})

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ch, err := e.ProcessTurnStream(context.Background(), uuid.New(), "look around")
	if err != nil {
		t.Fatalf("ProcessTurnStream() error = %v", err)
	}

	var events []StreamEvent
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != "status" {
		t.Fatalf("expected initial status event, got %q", events[0].Type)
	}
	if events[0].Status == nil || events[0].Status.Stage != "gathering" {
		t.Fatalf("expected gathering status event, got %+v", events[0].Status)
	}
	if events[1].Type != "error" {
		t.Fatalf("expected error event, got %q", events[1].Type)
	}
	if events[1].Err == nil {
		t.Fatal("expected non-nil Err on error event")
	}
}

func TestProcessTurnStream_ChannelCloses(t *testing.T) {
	e, err := New(nil, &testProvider{}, config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{ContextTokenBudget: 8000}})

	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ch, err := e.ProcessTurnStream(context.Background(), uuid.New(), "test")
	if err != nil {
		t.Fatalf("ProcessTurnStream() error = %v", err)
	}

	// Channel must close (no goroutine leak).
	for range ch {
	}
	// If we get here, channel closed successfully.
}

func TestProcessTurnStream_StatusIsolation(t *testing.T) {
	provider := &echoStatusProvider{}
	reg := tools.NewRegistry()
	e := &Engine{
		logger:    slog.Default(),
		state:     &stubStateManager{},
		assembler: assembly.NewContextAssembler(reg),
		processor: NewTurnProcessor(provider, reg, tools.NewValidator(reg), nil),
	}

	var mu sync.Mutex
	seen := map[string][]string{}
	record := func(key string, stages []string) {
		mu.Lock()
		defer mu.Unlock()
		seen[key] = append([]string(nil), stages...)
	}

	ch1, err := e.ProcessTurnStream(context.Background(), uuid.New(), "look left")
	if err != nil {
		t.Fatalf("ProcessTurnStream() error = %v", err)
	}
	ch2, err := e.ProcessTurnStream(context.Background(), uuid.New(), "look right")
	if err != nil {
		t.Fatalf("ProcessTurnStream() error = %v", err)
	}

	wait := func(ch <-chan StreamEvent) ([]string, string, error) {
		stages := []string{}
		result := ""
		var streamErr error
		deadline := time.After(5 * time.Second)
		for ev := range ch {
			if ev.Type == "status" && ev.Status != nil {
				stages = append(stages, ev.Status.Stage)
			}
			if ev.Type == "result" && ev.Result != nil {
				result = ev.Result.Narrative
			}
			if ev.Type == "error" && ev.Err != nil {
				streamErr = ev.Err
			}
			select {
			case <-deadline:
				t.Fatal("timed out waiting for stream events")
			default:
			}
		}
		return stages, result, streamErr
	}

	stages1, result1, err1 := wait(ch1)
	stages2, result2, err2 := wait(ch2)
	record("one", stages1)
	record("two", stages2)
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected stream errors: %v, %v", err1, err2)
	}
	if len(stages1) == 0 || len(stages2) == 0 {
		t.Fatalf("expected status events in both streams, got %v and %v", stages1, stages2)
	}
	if stages1[0] != "gathering" || stages2[0] != "gathering" {
		t.Fatalf("expected gathering as first status in each stream, got %v and %v", stages1, stages2)
	}
	if result1 != "done" || result2 != "done" {
		t.Fatalf("expected per-stream results to remain isolated, got %q and %q", result1, result2)
	}
	if got := seen["one"]; len(got) == 0 || got[0] != "gathering" {
		t.Fatalf("expected first stream to keep its own status sequence, got %v", got)
	}
}

// stubSearcher implements assembly.MemoryRetriever for testing.
type stubSearcher struct{}

func (s *stubSearcher) SearchSimilar(_ context.Context, _ uuid.UUID, _ string, _ int) ([]memory.MemoryResult, error) {
	return nil, nil
}

func TestWithTier3Retriever_SetsField(t *testing.T) {
	retriever := assembly.NewTier3Retriever(&stubSearcher{}, 5, nil)
	e, err := New(nil, &testProvider{}, config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{ContextTokenBudget: 8000}},
		WithTier3Retriever(retriever),
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if e.tier3 == nil {
		t.Fatal("expected tier3 to be non-nil after WithTier3Retriever")
	}
}

func TestNewWithoutTier3Retriever_NilField(t *testing.T) {
	e, err := New(nil, &testProvider{}, config.LLMConfig{Provider: "ollama", Ollama: config.OllamaConfig{ContextTokenBudget: 8000}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if e.tier3 != nil {
		t.Fatal("expected tier3 to be nil without WithTier3Retriever")
	}
}
