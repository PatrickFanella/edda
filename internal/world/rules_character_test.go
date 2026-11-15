package world

import (
	"context"
	"errors"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"reflect"
	"testing"
)

func TestRulesInterviewCorrectsAndResumesAfterProviderFailure(t *testing.T) {
	for _, class := range []string{"Fighter", "Cleric"} {
		t.Run(class, func(t *testing.T) {
			background := "Soldier"
			if class == "Cleric" {
				background = "Acolyte"
			}
			args := func(class, background string) map[string]any {
				return map[string]any{"name": "Mira", "concept": "Human " + class, "background": background, "personality": "Patient", "motivations": []any{"Protect the town"}, "strengths": []any{"Determination"}, "weaknesses": []any{"Pride"}, "rules_choices": map[string]any{"origin": "Human", "class": class, "background": background}}
			}
			response := func(a map[string]any) *llm.Response {
				return &llm.Response{ToolCalls: []llm.ToolCall{{Name: extractCharacterToolName, Arguments: a}}}
			}
			invented := args(class, background)
			invented["abilities"] = []any{"Teleport at will"}
			missing := args(class, background)
			delete(missing, "rules_choices")
			provider := newScriptedProvider(t, scriptedResponse{resp: &llm.Response{Content: "Choose a supported build."}}, scriptedResponse{resp: response(args("Wizard", "Sage"))}, scriptedResponse{resp: response(invented)}, scriptedResponse{resp: response(missing)}, scriptedResponse{err: errors.New("provider unavailable")}, scriptedResponse{resp: response(args(class, background))})
			iv := NewRulesCharacterInterviewer(provider, nil, domain.RulesVersionSRD521)
			if _, err := iv.Start(context.Background()); err != nil {
				t.Fatal(err)
			}
			for range 3 {
				r, err := iv.Step(context.Background(), "Propose my choices")
				if err != nil {
					t.Fatal(err)
				}
				if r.Done || r.Profile != nil {
					t.Fatal("invalid choices completed")
				}
			}
			before := iv.History()
			if _, err := iv.Step(context.Background(), "Use Human "+class); err == nil {
				t.Fatal("provider failure lost")
			}
			if !reflect.DeepEqual(before, iv.History()) {
				t.Fatal("failed turn polluted resumable transcript")
			}
			r, err := iv.Step(context.Background(), "Use Human "+class)
			if err != nil {
				t.Fatal(err)
			}
			if !r.Done || r.Profile == nil || r.Profile.RulesChoices == nil {
				t.Fatalf("result=%+v", r)
			}
			interviewBuild, err := ValidateRulesCharacter(r.Profile)
			if err != nil {
				t.Fatal(err)
			}
			guidedBuild, err := ValidateRulesCharacter(&CharacterProfile{Name: "Mira", Concept: "Human " + class, Background: background})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(interviewBuild, guidedBuild) {
				t.Fatal("interview and guided mechanics diverge")
			}
		})
	}
}

func TestRulesCharacterRejectsMismatchedStructuredChoices(t *testing.T) {
	if _, err := ValidateRulesCharacter(&CharacterProfile{Name: "Mira", Concept: "Human Fighter", Background: "Soldier", RulesChoices: &RulesChoices{Origin: "Human", Class: "Cleric", Background: "Acolyte"}}); err == nil {
		t.Fatal("mismatched choices accepted")
	}
}
