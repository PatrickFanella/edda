package llm

import (
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
)

func TestNewRosterRoutesProfilesAndFallsBackToDefault(t *testing.T) {
	cfg := config.Config{LLM: config.LLMConfig{
		Provider:   "openrouter",
		OpenRouter: config.OpenRouterConfig{APIKey: "sk-or-test", Model: "default-model", ContextTokenBudget: 8000},
		Profiles: map[string]config.LLMProfileConfig{
			"post-turn": {
				Provider:           "openrouter",
				Model:              "qwen/qwen3-235b-a22b-2507",
				ContextTokenBudget: 12000,
				ToolSupport:        true,
			},
		},
		Routes: config.LLMRoutesConfig{PostTurnState: "post-turn"},
	}}

	roster, err := NewRoster(cfg)
	if err != nil {
		t.Fatalf("NewRoster() error = %v", err)
	}

	postTurn := roster.Info(FlowPostTurnState)
	if postTurn.ProfileName != "post-turn" {
		t.Fatalf("expected post-turn profile, got %q", postTurn.ProfileName)
	}
	if postTurn.Model != "qwen/qwen3-235b-a22b-2507" {
		t.Fatalf("unexpected routed model: %q", postTurn.Model)
	}
	if postTurn.ContextTokenBudget != 12000 {
		t.Fatalf("unexpected routed context token budget: %d", postTurn.ContextTokenBudget)
	}
	if postTurn.Provider == nil {
		t.Fatal("expected routed provider")
	}

	choice := roster.Info(FlowChoiceFallback)
	if choice.ProfileName != "default" {
		t.Fatalf("expected choice fallback to use default profile, got %q", choice.ProfileName)
	}
	if choice.Model != "default-model" {
		t.Fatalf("unexpected fallback model: %q", choice.Model)
	}
}

func TestNewRosterProfileCanOverrideOnlyModel(t *testing.T) {
	cfg := config.Config{LLM: config.LLMConfig{
		Provider:   "openrouter",
		OpenRouter: config.OpenRouterConfig{APIKey: "sk-or-test", Model: "default-model", ContextTokenBudget: 8000},
		Profiles: map[string]config.LLMProfileConfig{
			"choice": {Model: "mistralai/mistral-small-3.2-24b-instruct"},
		},
		Routes: config.LLMRoutesConfig{ChoiceFallback: "choice"},
	}}

	roster, err := NewRoster(cfg)
	if err != nil {
		t.Fatalf("NewRoster() error = %v", err)
	}
	routed := roster.Info(FlowChoiceFallback)
	if routed.ProviderName != "openrouter" {
		t.Fatalf("expected inherited provider openrouter, got %q", routed.ProviderName)
	}
	if routed.Model != "mistralai/mistral-small-3.2-24b-instruct" {
		t.Fatalf("unexpected model: %q", routed.Model)
	}
	if routed.ContextTokenBudget != 8000 {
		t.Fatalf("expected inherited budget, got %d", routed.ContextTokenBudget)
	}
}

func TestRosterGMProfileTurnsOnClaudeCodeSkills(t *testing.T) {
	cfg := config.Config{LLM: config.LLMConfig{
		Provider:   "claudecode",
		ClaudeCode: config.ClaudeCodeConfig{Endpoint: "http://bridge", Token: "tok"},
		Profiles:   map[string]config.LLMProfileConfig{"gm": {Provider: "claudecode", Skills: true, ToolSupport: true}},
		Routes:     config.LLMRoutesConfig{GMTurn: "gm"},
	}}
	roster, err := NewRoster(cfg)
	if err != nil {
		t.Fatal(err)
	}
	gm, ok := roster.Provider(FlowGMTurn).(*ClaudeCodeClient)
	if !ok || !gm.skills {
		t.Fatalf("gm provider = %T skills=%v, want *ClaudeCodeClient with skills", roster.Provider(FlowGMTurn), ok && gm.skills)
	}
	def, ok := roster.Provider(FlowPostTurnState).(*ClaudeCodeClient)
	if !ok || def.skills {
		t.Fatalf("post-turn provider should be the default claudecode client without skills")
	}
}
