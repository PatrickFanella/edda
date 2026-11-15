package game

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"github.com/google/uuid"
)

func previewFixture(t *testing.T) (domain.Campaign, domain.PlayerCharacter) {
	t.Helper()
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	c := domain.Campaign{ID: uuid.New(), RulesVersion: domain.RulesVersionSRD521, Status: domain.CampaignStatusActive}
	pc := domain.PlayerCharacter{ID: uuid.New(), CampaignID: c.ID, Name: "Mira", Stats: stats, HP: 9, MaxHP: 12, Level: 1, Status: "active"}
	return c, pc
}
func TestActionPreviewSupportedChecks(t *testing.T) {
	c, pc := previewFixture(t)
	before, _ := json.Marshal(pc)
	for _, tc := range []struct {
		kind, ability, skill string
		modifier, targets    int
	}{
		{"ability-check", "strength", "", 3, 0}, {"ability-check", "", "athletics", 5, 0}, {"saving-throw", "constitution", "", 4, 1},
	} {
		t.Run(tc.kind+tc.ability+tc.skill, func(t *testing.T) {
			intent := ActionIntent{ActorID: pc.ID, Kind: tc.kind, Ability: tc.ability, Skill: tc.skill}
			got, err := PrepareCharacterAction(context.Background(), c, pc, intent)
			if err != nil {
				t.Fatal(err)
			}
			if !got.Available || got.Modifier.Value != tc.modifier || len(got.Targets) != tc.targets || len(got.Costs) != 0 || got.ExpectedRevision == "" || len(got.SourceIDs) == 0 || !reflect.DeepEqual(got.Options, []string{"ordinary"}) {
				t.Fatalf("bad preview %+v", got)
			}
			again, err := PrepareCharacterAction(context.Background(), c, pc, intent)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, again) {
				t.Fatal("preview not deterministic")
			}
			if got.Intent.Skill != "" && got.Intent.Ability != "strength" {
				t.Fatal("skill not normalized")
			}
		})
	}
	after, _ := json.Marshal(pc)
	if string(before) != string(after) {
		t.Fatal("preview changed actor")
	}
}
func TestActionPreviewUnavailable(t *testing.T) {
	for _, name := range []string{"foreign actor", "legacy", "build", "entry", "kind", "ability", "pairing", "paused", "inactive", "unconscious"} {
		t.Run(name, func(t *testing.T) {
			c, pc := previewFixture(t)
			intent := ActionIntent{ActorID: pc.ID, Kind: "ability-check", Ability: "strength"}
			switch name {
			case "foreign actor":
				intent.ActorID = uuid.New()
			case "legacy":
				c.RulesVersion = domain.RulesVersionLegacy
			case "build":
				pc.Stats = []byte(`{"srd521":{"class":"secret-build"}}`)
			case "entry":
				intent.EntryID = "attack"
			case "kind":
				intent.Kind = "attack"
			case "ability":
				intent.Ability = "secret-ability"
			case "pairing":
				intent.Skill = "perception"
			case "paused":
				c.Status = domain.CampaignStatusPaused
			case "inactive":
				pc.Status = "dead"
			case "unconscious":
				pc.HP = 0
			}
			got, err := PrepareCharacterAction(context.Background(), c, pc, intent)
			if err != nil {
				t.Fatal(err)
			}
			if got.Available || got.Reason == "" || got.Modifier != nil || len(got.Targets) != 0 {
				t.Fatalf("bad unavailable %+v", got)
			}
		})
	}
}
func TestActionPreviewRevisionAndCancellation(t *testing.T) {
	c, pc := previewFixture(t)
	intent := ActionIntent{ActorID: pc.ID, Kind: srd521.AbilityCheck, Ability: "strength"}
	first, err := PrepareCharacterAction(context.Background(), c, pc, intent)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"hp", "resources", "campaign", "location", "revision"} {
		t.Run(name, func(t *testing.T) {
			changedC, changedPC := c, pc
			switch name {
			case "hp":
				changedPC.HP--
			case "resources":
				changedPC.Stats = append([]byte(nil), pc.Stats...)
				var raw map[string]any
				if err := json.Unmarshal(changedPC.Stats, &raw); err != nil {
					t.Fatal(err)
				}
				raw["srd521"].(map[string]any)["second_wind"] = 0
				changedPC.Stats, _ = json.Marshal(raw)
			case "campaign":
				changedC.Status = domain.CampaignStatusPaused
			case "location":
				id := uuid.New()
				changedPC.CurrentLocationID = &id
			case "revision":
				changedPC.UpdatedAt = time.Now()
			}
			got, err := PrepareCharacterAction(context.Background(), changedC, changedPC, intent)
			if err != nil {
				t.Fatal(err)
			}
			if first.ExpectedRevision == got.ExpectedRevision {
				t.Fatal("stale actor revision")
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := PrepareCharacterAction(ctx, c, pc, intent); err == nil {
		t.Fatal("ignored cancellation")
	}
}
