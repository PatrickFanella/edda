package prompt

import (
	"strings"
	"testing"
)

func TestGameMasterPromptNonEmpty(t *testing.T) {
	if strings.TrimSpace(GameMaster) == "" {
		t.Fatal("GameMaster prompt must not be empty")
	}
}

func TestGameMasterPromptSections(t *testing.T) {
	requiredSections := []struct {
		name    string
		heading string
	}{
		{"narrative voice", "NARRATIVE VOICE AND TONE"},
		{"tool usage", "TOOL USAGE GUIDELINES"},
		{"choice presentation", "CHOICE PRESENTATION"},
		{"combat guidelines", "COMBAT GUIDELINES"},
		{"pacing guidelines", "PACING GUIDELINES"},
		{"consistency rules", "CONSISTENCY RULES"},
		{"dice rolling", "DICE ROLLING AND SKILL CHECKS"},
		{"new content", "INTRODUCING NEW CONTENT"},
	}

	for _, s := range requiredSections {
		t.Run(s.name, func(t *testing.T) {
			if !strings.Contains(GameMaster, s.heading) {
				t.Fatalf("GameMaster prompt missing section %q", s.heading)
			}
		})
	}
}

func TestGameMasterPromptToolReferences(t *testing.T) {
	tools := []string{"skill_check", "create_language", "roll_dice", "create_quest", "update_player_hp"}
	for _, tool := range tools {
		if !strings.Contains(GameMaster, tool) {
			t.Fatalf("GameMaster prompt must reference tool %q", tool)
		}
	}
}

func TestGameMasterPromptDurableStateClaims(t *testing.T) {
	phrases := []string{
		"Durable state claims",
		"must call the matching state tool before claiming it in prose",
		"create_location with move_player_here=true",
		"Revealing or creating a location without movement is not enough",
		"Do not claim an item was added unless you called add_item or create_item",
		"Do not claim an item was removed unless you called remove_item",
		"Do not say combat has begun unless you called initiate_combat",
		"do not say combat is over unless you called resolve_combat",
		"If the needed tool is unavailable",
		"provisional",
	}
	for _, phrase := range phrases {
		if !strings.Contains(GameMaster, phrase) {
			t.Fatalf("GameMaster prompt must include durable-state guidance phrase %q", phrase)
		}
	}
}

func TestGameMasterPromptQuestProgressGuidance(t *testing.T) {
	phrases := []string{
		"If an active quest already exists",
		"use update_quest or complete_objective",
		"Do not create a duplicate quest",
	}
	for _, phrase := range phrases {
		if !strings.Contains(GameMaster, phrase) {
			t.Fatalf("GameMaster prompt must include quest progress guidance phrase %q", phrase)
		}
	}
}

func TestGameMasterPromptNoRepeatDurableMutations(t *testing.T) {
	phrases := []string{
		"Only call durable mutation tools for the current player input",
		"Do not reapply the same prior event or tool result",
		"ongoing effect",
		"poison, bleeding, burning, timers, pursuit, or forced movement",
	}
	for _, phrase := range phrases {
		if !strings.Contains(GameMaster, phrase) {
			t.Fatalf("GameMaster prompt must include no-repeat guidance phrase %q", phrase)
		}
	}
}

func TestGameMasterPromptExplicitCombatGuidance(t *testing.T) {
	phrases := []string{
		"hostile creature or drone actively blocks the route",
		"call initiate_combat before narrating that combat has begun",
	}
	for _, phrase := range phrases {
		if !strings.Contains(GameMaster, phrase) {
			t.Fatalf("GameMaster prompt must include explicit combat guidance phrase %q", phrase)
		}
	}
}

func TestGameMasterPromptCombatStartGuidance(t *testing.T) {
	phrases := []string{
		"hostile creature or drone",
		"locks onto you",
		"call initiate_combat before narrating that combat has begun",
		"keep it as tension rather than starting combat",
	}
	for _, phrase := range phrases {
		if !strings.Contains(GameMaster, phrase) {
			t.Fatalf("GameMaster prompt must include combat guidance phrase %q", phrase)
		}
	}
}

func TestGameMasterPromptNarrativeVoice(t *testing.T) {
	if !strings.Contains(GameMaster, "second person") {
		t.Fatal("GameMaster prompt must mention second person narrative voice")
	}
	if !strings.Contains(GameMaster, "third person") {
		t.Fatal("GameMaster prompt must mention third person narrative voice")
	}
}

func TestGameMasterPromptChoiceGuidelines(t *testing.T) {
	if !strings.Contains(GameMaster, "3 and 5") {
		t.Fatal("GameMaster prompt must specify offering 3 to 5 choices")
	}
	if !strings.Contains(GameMaster, "never limited to the options above") {
		t.Fatal("GameMaster prompt must signal that free input is welcome")
	}
}

func TestGameMasterPromptDifficultyClasses(t *testing.T) {
	dcs := []string{"DC 10", "DC 15", "DC 20", "DC 25"}
	for _, dc := range dcs {
		if !strings.Contains(GameMaster, dc) {
			t.Fatalf("GameMaster prompt must reference %s", dc)
		}
	}
}

func TestGameMasterPromptConsistency(t *testing.T) {
	keywords := []string{
		"established facts",
		"NPC states",
		"player choices",
	}
	for _, kw := range keywords {
		if !strings.Contains(GameMaster, kw) {
			t.Fatalf("GameMaster prompt must reference %q for consistency rules", kw)
		}
	}
}

func TestGameMasterPromptPacingPillars(t *testing.T) {
	pillars := []string{"Exposition", "Action", "Dialogue"}
	for _, p := range pillars {
		if !strings.Contains(GameMaster, p) {
			t.Fatalf("GameMaster prompt must reference pacing pillar %q", p)
		}
	}
}
