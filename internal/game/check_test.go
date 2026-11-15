package game_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
)

type fixtureDice struct{ rolls []int }

func (d *fixtureDice) Intn(_ int) int {
	v := d.rolls[0]
	d.rolls = d.rolls[1:]
	return v
}

func fixtureCharacter(t *testing.T) domain.PlayerCharacter {
	t.Helper()
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	return domain.PlayerCharacter{Stats: stats, Level: 1, HP: 9, MaxHP: 12, Experience: 125}
}

// This golden file is also consumed by the frontend behavior suite. The only
// arithmetic path generating its receipts is the shared game/rules boundary.
func TestCheckSheetFixture(t *testing.T) {
	pc := fixtureCharacter(t)
	sheet, err := srd521.Project(pc.Stats, pc.Level, pc.HP, pc.MaxHP, pc.Experience)
	if err != nil {
		t.Fatal(err)
	}
	fixture := struct {
		Sheet    *srd521.CharacterSheet `json:"sheet"`
		Receipts []api.CheckReceipt     `json:"receipts"`
	}{Sheet: sheet}
	for _, tc := range []struct {
		check srd521.Check
		dc    srd521.Difficulty
		rolls []int
	}{
		{srd521.Check{Kind: srd521.AbilityCheck, Ability: "strength", Skill: "athletics"}, srd521.Difficulty{Value: 15, PlayerKnown: true, OutcomeKnown: true}, []int{9}},
		{srd521.Check{Kind: srd521.AbilityCheck, Ability: "strength", Skill: "athletics", Advantage: true}, srd521.Difficulty{Value: 17, OutcomeKnown: true}, []int{2, 14}},
		{srd521.Check{Kind: srd521.SavingThrow, Ability: "constitution", Disadvantage: true}, srd521.Difficulty{Value: 12, PlayerKnown: true, OutcomeKnown: true}, []int{16, 3}},
		{srd521.Check{Kind: srd521.AbilityCheck, Ability: "wisdom", Skill: "perception"}, srd521.Difficulty{Value: 17}, []int{9}},
	} {
		resolution, err := game.ResolveCharacterCheck(context.Background(), domain.RulesVersionSRD521, pc, tc.check, tc.dc, &fixtureDice{tc.rolls})
		if err != nil {
			t.Fatal(err)
		}
		result := tools.SRDCheckResult(*resolution)
		receipt, ok := result.Data["receipt"].(api.CheckReceipt)
		if !ok || !result.Success || len(result.Data) != 1 || result.Narrative != "" {
			t.Fatalf("tool adapter exposed non-receipt context: %+v", result)
		}
		fixture.Receipts = append(fixture.Receipts, receipt)
	}
	got, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	const path = "../../frontend/tests/fixtures/srd-check-sheet.json"
	if os.Getenv("EDDA_UPDATE_CHECK_FIXTURE") == "1" {
		if err := os.WriteFile(path, got, 0644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatal("sheet/check fixture drifted; review the rules change before regenerating with EDDA_UPDATE_CHECK_FIXTURE=1")
	}
}

func TestCharacterCheckGuards(t *testing.T) {
	pc := fixtureCharacter(t)
	check := srd521.Check{Kind: srd521.AbilityCheck, Ability: "strength"}
	for _, version := range []domain.RulesVersion{"", domain.RulesVersionLegacy, "future"} {
		if _, err := game.ResolveCharacterCheck(context.Background(), version, pc, check, srd521.Difficulty{}, &fixtureDice{}); err == nil {
			t.Fatalf("accepted campaign version %s", version)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := game.ResolveCharacterCheck(ctx, domain.RulesVersionSRD521, pc, check, srd521.Difficulty{}, &fixtureDice{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	pc.Stats = []byte(`{"strength":16}`)
	if _, err := game.ResolveCharacterCheck(context.Background(), domain.RulesVersionSRD521, pc, check, srd521.Difficulty{}, &fixtureDice{}); err == nil {
		t.Fatal("legacy character accepted as SRD")
	}
}
