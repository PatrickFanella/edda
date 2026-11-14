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

func TestRosterSkillsProfileWrapsProvider(t *testing.T) {
	cfg := config.Config{LLM: config.LLMConfig{
		Provider:   "claudecode",
		ClaudeCode: config.ClaudeCodeConfig{Endpoint: "http://bridge", Token: "tok"},
		SkillsDir:  "../../skills",
		Profiles:   map[string]config.LLMProfileConfig{"gm": {Skills: true, ToolSupport: true}},
		Routes:     config.LLMRoutesConfig{GMTurn: "gm"},
	}}
	roster, err := NewRoster(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := roster.Provider(FlowGMTurn).(*SkillProvider); !ok {
		t.Fatalf("gm provider = %T, want *SkillProvider", roster.Provider(FlowGMTurn))
	}
	if _, ok := roster.Provider(FlowPostTurnState).(*ClaudeCodeClient); !ok {
		t.Fatalf("other flows should not get skills, got %T", roster.Provider(FlowPostTurnState))
	}

	cfg.LLM.SkillsDir = t.TempDir()
	roster, err = NewRoster(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := roster.Provider(FlowGMTurn).(*ClaudeCodeClient); !ok {
		t.Fatalf("an empty skills folder should leave the provider unwrapped, got %T", roster.Provider(FlowGMTurn))
	}
}
