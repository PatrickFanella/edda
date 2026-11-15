//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

func TestIntegrationFighterAdvancement(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, choice := range []string{"fixed", "roll"} {
		t.Run(choice, func(t *testing.T) {
			creation, err := srd521.SupportedBuildForConcept("Human Fighter", "Soldier")
			must(err)
			stats, err := creation.Stats()
			must(err)
			c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Advancement", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
			must(err)
			t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
			pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
			must(err)
			campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
			prepare := func() *game.ActionPreview {
				t.Helper()
				c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
				must(err)
				a, err := game.PrepareStoredAction(ctx, testPool, c, p, game.ActionIntent{ActorID: actor, Kind: srd521.AdvanceLevel, EntryID: choice}, false)
				must(err)
				return a.Preview
			}
			if prepare().Available {
				t.Fatal("advanced without earned XP")
			}
			_, err = testPool.Exec(ctx, `UPDATE player_characters SET experience=300,stats=stats||'{"custom":"keep"}'::jsonb WHERE id=$1`, pc.ID)
			must(err)
			preview := prepare()
			if !preview.Available {
				t.Fatalf("preview=%+v", preview)
			}
			dice := &lifecycleIntegrationDice{values: []int{4}}
			newEngine := func() *engine.Engine {
				e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
				must(err)
				return e
			}
			request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: preview.Intent}
			// A failed history write must roll the complete transition back.
			_, err = testPool.Exec(ctx, `ALTER TABLE session_logs ADD CONSTRAINT advancement_failure CHECK(player_input NOT LIKE 'Confirm Fighter level 2%')`)
			must(err)
			t.Cleanup(func() {
				_, _ = testPool.Exec(ctx, `ALTER TABLE session_logs DROP CONSTRAINT IF EXISTS advancement_failure`)
			})
			if _, err := newEngine().ConfirmAction(ctx, campaign, request); err == nil {
				t.Fatal("expected history failure")
			}
			failed, err := q.GetPlayerCharacterByID(ctx, pc.ID)
			must(err)
			if failed.Level != 1 || failed.MaxHp != 12 {
				t.Fatal("partial advancement")
			}
			_, err = testPool.Exec(ctx, `ALTER TABLE session_logs DROP CONSTRAINT advancement_failure`)
			must(err)
			failedCalls := dice.calls
			dice.values = []int{4}
			first, err := newEngine().ConfirmAction(ctx, campaign, request)
			must(err)
			again, err := newEngine().ConfirmAction(ctx, campaign, request)
			must(err)
			a, err := json.Marshal(first)
			must(err)
			b, err := json.Marshal(again)
			must(err)
			var decodedA, decodedB any
			must(json.Unmarshal(a, &decodedA))
			must(json.Unmarshal(b, &decodedB))
			if !reflect.DeepEqual(decodedA, decodedB) {
				t.Fatal("receipt replay changed")
			}
			reloaded, err := q.GetPlayerCharacterByID(ctx, pc.ID)
			must(err)
			expected, calls := 20, 0
			if choice == "roll" {
				expected, calls = 19, 1
			}
			if reloaded.Level != 2 || int(reloaded.MaxHp) != expected || reloaded.Hp != 9 || dice.calls-failedCalls != calls {
				t.Fatalf("pc=%+v rolls=%d", reloaded, dice.calls)
			}
			var saved map[string]any
			must(json.Unmarshal(reloaded.Stats, &saved))
			if saved["custom"] != "keep" {
				t.Fatal("unrelated stats changed")
			}
			if _, err := srd521.Project(reloaded.Stats, 2, 9, expected, 300); err != nil {
				t.Fatal(err)
			}
			stale := request
			stale.ActionID = uuid.New()
			if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
				t.Fatalf("repeat=%v", err)
			}
			if prepare().Available {
				t.Fatal("level 2 advanced again")
			}
		})
	}
}
