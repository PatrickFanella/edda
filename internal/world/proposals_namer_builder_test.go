package world

import (
	"context"
	"slices"
	"testing"

	"github.com/PatrickFanella/game-master/internal/llm"
)

func validProposalResponse() proposalLLMResponse {
	return proposalLLMResponse{
		Proposals: []proposalEntry{
			{
				Name:                "Ashes of Blackfen",
				Summary:             "Survivors navigate a plague-ridden marsh where rival houses weaponize old curses.",
				Genre:               "dark fantasy",
				Tone:                "grim",
				Themes:              []string{"survival", "betrayal"},
				WorldType:           "haunted marsh kingdom",
				DangerLevel:         "high",
				PoliticalComplexity: "complex",
			},
			{
				Name:                "Clockwork Ember",
				Summary:             "A smoke-choked city teeters between guild war and a machine cult awakening below it.",
				Genre:               "steampunk",
				Tone:                "tense",
				Themes:              []string{"ambition", "class conflict"},
				WorldType:           "industrial city-state",
				DangerLevel:         "moderate",
				PoliticalComplexity: "intricate",
			},
		},
	}
}

func assertCampaignProposalMatches(t *testing.T, got CampaignProposal, want proposalEntry) {
	t.Helper()

	if got.Name != want.Name {
		t.Errorf("name: got %q, want %q", got.Name, want.Name)
	}
	if got.Summary != want.Summary {
		t.Errorf("summary: got %q, want %q", got.Summary, want.Summary)
	}
	if got.Profile.Genre != want.Genre {
		t.Errorf("profile.genre: got %q, want %q", got.Profile.Genre, want.Genre)
	}
	if got.Profile.Tone != want.Tone {
		t.Errorf("profile.tone: got %q, want %q", got.Profile.Tone, want.Tone)
	}
	if !slices.Equal(got.Profile.Themes, want.Themes) {
		t.Errorf("profile.themes: got %v, want %v", got.Profile.Themes, want.Themes)
	}
	if got.Profile.WorldType != want.WorldType {
		t.Errorf("profile.world_type: got %q, want %q", got.Profile.WorldType, want.WorldType)
	}
	if got.Profile.DangerLevel != want.DangerLevel {
		t.Errorf("profile.danger_level: got %q, want %q", got.Profile.DangerLevel, want.DangerLevel)
	}
	if got.Profile.PoliticalComplexity != want.PoliticalComplexity {
		t.Errorf("profile.political_complexity: got %q, want %q", got.Profile.PoliticalComplexity, want.PoliticalComplexity)
	}
}

func TestProposalGenerator_Success(t *testing.T) {
	resp := validProposalResponse()
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: mustJSON(t, resp)}})
	gen := NewProposalGenerator(provider)

	got, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != len(resp.Proposals) {
		t.Fatalf("proposal count: got %d, want %d", len(got), len(resp.Proposals))
	}

	assertCampaignProposalMatches(t, got[0], resp.Proposals[0])
	assertCampaignProposalMatches(t, got[1], resp.Proposals[1])
}

func TestProposalGenerator_MarkdownFences(t *testing.T) {
	resp := validProposalResponse()
	wrapped := "```json\n" + mustJSON(t, resp) + "\n```"
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: wrapped}})
	gen := NewProposalGenerator(provider)

	got, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(got) != len(resp.Proposals) {
		t.Fatalf("proposal count: got %d, want %d", len(got), len(resp.Proposals))
	}
	assertCampaignProposalMatches(t, got[0], resp.Proposals[0])
}

func TestProposalGenerator_LeadingTextAndTrailingText(t *testing.T) {
	resp := validProposalResponse()
	raw := mustJSON(t, resp)
	content := "Here are three options for your campaign:\n\n" + raw + "\n\nPick the one you like best."
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: content}})
	gen := NewProposalGenerator(provider)

	got, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(resp.Proposals) {
		t.Fatalf("proposal count: got %d, want %d", len(got), len(resp.Proposals))
	}
	assertCampaignProposalMatches(t, got[0], resp.Proposals[0])
}

func TestProposalGenerator_NestedProfilePayload(t *testing.T) {
	content := `{"proposals":[{"name":"Veil of Salt","summary":"A storm-lashed island mystery.","profile":{"genre":"fantasy","tone":"mysterious","themes":["secrets","loss"],"world_type":"island archipelago","danger_level":"high","political_complexity":"complex"}}]}`
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: content}})
	gen := NewProposalGenerator(provider)

	got, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("proposal count: got %d, want 1", len(got))
	}
	if got[0].Profile.WorldType != "island archipelago" {
		t.Fatalf("world type: got %q, want %q", got[0].Profile.WorldType, "island archipelago")
	}
}

func TestProposalGenerator_EmptyLLMResponse(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "   "}})
	gen := NewProposalGenerator(provider)

	_, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err == nil {
		t.Fatal("expected error for empty LLM response")
	}
}

func TestProposalGenerator_MalformedJSON(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "not json"}})
	gen := NewProposalGenerator(provider)

	_, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestProposalGenerator_ZeroProposals(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: mustJSON(t, proposalLLMResponse{})}})
	gen := NewProposalGenerator(provider)

	_, err := gen.Generate(context.Background(), "fantasy", "wilderness", "grim")
	if err == nil {
		t.Fatal("expected error for zero proposals")
	}
}

func TestGenerateCampaignName_Success(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: `{"name":"Shadows of Ironhold"}`}})

	got, err := GenerateCampaignName(context.Background(), provider, testProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Shadows of Ironhold" {
		t.Fatalf("name: got %q, want %q", got, "Shadows of Ironhold")
	}
}

func TestGenerateCampaignName_MarkdownFences(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "```json\n{\"name\":\"Shadows of Ironhold\"}\n```"}})

	got, err := GenerateCampaignName(context.Background(), provider, testProfile())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Shadows of Ironhold" {
		t.Fatalf("name: got %q, want %q", got, "Shadows of Ironhold")
	}
}

func TestGenerateCampaignName_NilProfile(t *testing.T) {
	provider := newScriptedProvider(t)

	_, err := GenerateCampaignName(context.Background(), provider, nil)
	if err == nil {
		t.Fatal("expected error for nil profile")
	}
}

func TestGenerateCampaignName_EmptyResponse(t *testing.T) {
	provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "\n\t "}})

	_, err := GenerateCampaignName(context.Background(), provider, testProfile())
	if err == nil {
		t.Fatal("expected error for empty response")
	}
}

func TestGenerateCampaignName_MissingOrEmptyName(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "missing field", content: `{}`},
		{name: "blank field", content: `{"name":"   "}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: tt.content}})

			_, err := GenerateCampaignName(context.Background(), provider, testProfile())
			if err == nil {
				t.Fatal("expected error for missing or empty name")
			}
		})
	}
}

func TestBuildCharacterFromAttributes_KnownValues(t *testing.T) {
	attrs := CharacterAttributes{
		Name:       "Aelar",
		Race:       "Elf",
		Class:      "Wizard",
		Background: "Noble",
		Alignment:  "Chaotic Good",
		Traits:     []string{"Curious", "Patient"},
	}

	got := BuildCharacterFromAttributes(attrs)
	if got == nil {
		t.Fatal("expected character profile")
	}

	if got.Name != "Aelar" {
		t.Errorf("name: got %q, want %q", got.Name, "Aelar")
	}
	if got.Concept != "elf wizard" {
		t.Errorf("concept: got %q, want %q", got.Concept, "elf wizard")
	}
	if got.Background != "Noble" {
		t.Errorf("background: got %q, want %q", got.Background, "Noble")
	}
	if got.Personality != "Chaotic Good; Curious; Patient" {
		t.Errorf("personality: got %q, want %q", got.Personality, "Chaotic Good; Curious; Patient")
	}
	if !slices.Equal(got.Motivations, classMotivations["Wizard"]) {
		t.Errorf("motivations: got %v, want %v", got.Motivations, classMotivations["Wizard"])
	}

	wantStrengths := append(append([]string{}, raceStrengths["Elf"]...), classStrengths["Wizard"]...)
	if !slices.Equal(got.Strengths, wantStrengths) {
		t.Errorf("strengths: got %v, want %v", got.Strengths, wantStrengths)
	}
	if !slices.Equal(got.Weaknesses, backgroundWeaknesses["Noble"]) {
		t.Errorf("weaknesses: got %v, want %v", got.Weaknesses, backgroundWeaknesses["Noble"])
	}
}

func TestBuildCharacterFromAttributes_UnknownValuesFallback(t *testing.T) {
	attrs := CharacterAttributes{
		Name:       "Rook",
		Race:       "Goblin",
		Class:      "Artificer",
		Background: "Mercenary",
		Alignment:  "Neutral",
	}

	got := BuildCharacterFromAttributes(attrs)
	if got == nil {
		t.Fatal("expected character profile")
	}

	if got.Name != "Rook" {
		t.Errorf("name: got %q, want %q", got.Name, "Rook")
	}
	if got.Concept != "goblin artificer" {
		t.Errorf("concept: got %q, want %q", got.Concept, "goblin artificer")
	}
	if !slices.Equal(got.Motivations, defaultMotivations) {
		t.Errorf("motivations: got %v, want %v", got.Motivations, defaultMotivations)
	}

	wantStrengths := append(append([]string{}, defaultStrengths...), defaultStrengths...)
	if !slices.Equal(got.Strengths, wantStrengths) {
		t.Errorf("strengths: got %v, want %v", got.Strengths, wantStrengths)
	}
	if !slices.Equal(got.Weaknesses, defaultWeaknesses) {
		t.Errorf("weaknesses: got %v, want %v", got.Weaknesses, defaultWeaknesses)
	}
	if len(got.Motivations) == 0 || len(got.Strengths) == 0 || len(got.Weaknesses) == 0 {
		t.Fatalf("expected non-empty fallback slices, got motivations=%v strengths=%v weaknesses=%v", got.Motivations, got.Strengths, got.Weaknesses)
	}
}

func TestBuildCharacterFromAttributes_EmptyTraitsDoesNotAppendSeparator(t *testing.T) {
	attrs := CharacterAttributes{
		Name:       "Sera",
		Race:       "Human",
		Class:      "Fighter",
		Background: "Soldier",
		Alignment:  "Lawful Neutral",
		Traits:     nil,
	}

	got := BuildCharacterFromAttributes(attrs)
	if got == nil {
		t.Fatal("expected character profile")
	}

	if got.Personality != "Lawful Neutral" {
		t.Errorf("personality: got %q, want %q", got.Personality, "Lawful Neutral")
	}
	if got.Concept != "human fighter" {
		t.Errorf("concept: got %q, want %q", got.Concept, "human fighter")
	}
}
