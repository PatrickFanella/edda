//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

type spatialDice struct{ calls, roll int }

func (d *spatialDice) Intn(n int) int {
	d.calls++
	if n == 20 {
		return d.roll
	}
	return 0
}

type spatialFixture struct {
	t                               *testing.T
	campaign, actor, target, weapon uuid.UUID
	dice                            *spatialDice
}

func newSpatialFixture(t *testing.T) *spatialFixture {
	t.Helper()
	f := &spatialFixture{t: t, dice: &spatialDice{roll: 18}}
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	f.must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Spatial fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	f.must(err)
	f.campaign = dbutil.FromPgtype(c.ID)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	f.must(err)
	f.actor = dbutil.FromPgtype(pc.ID)
	var location uuid.UUID
	f.must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name)VALUES($1,'Court')RETURNING id`, c.ID).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, location, pc.ID)
	f.must(err)
	f.must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,properties)VALUES($1,$2,'Longsword','weapon','common',true,'{"srd521_id":"longsword"}')RETURNING id`, c.ID, pc.ID).Scan(&f.weapon))
	f.must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id)VALUES($1,'Practice foe',100,$2)RETURNING id`, c.ID, location).Scan(&f.target))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved)VALUES($1,1,'Arrive','game_action','Practice foe waits.',ARRAY[$2]::uuid[])`, c.ID, f.target)
	f.must(err)
	f.facts(`"distance_feet":15,"line_of_sight":true,"cover":"none"`)
	return f
}
func (f *spatialFixture) must(err error) {
	f.t.Helper()
	if err != nil {
		f.t.Fatal(err)
	}
}
func (f *spatialFixture) facts(facts string) {
	f.t.Helper()
	props := fmt.Sprintf(`{"srd521_target":{"profile":"ordinary-spatial-fixture-v1","actor_id":"%s",%s},"secret":"unpublished villain"}`, f.actor, facts)
	_, err := testPool.Exec(context.Background(), `UPDATE npcs SET properties=$1 WHERE id=$2`, props, f.target)
	f.must(err)
}
func (f *spatialFixture) prepare(kind string) *game.ActionPreview {
	f.t.Helper()
	ctx := context.Background()
	c, pc, err := game.ReadActionSnapshot(ctx, testPool, f.campaign, f.actor, testUserID, false)
	f.must(err)
	intent := game.ActionIntent{ActorID: f.actor, Kind: kind}
	if kind == combat.EncounterStart || kind == srd521.MartialAttack {
		intent.EntryID = f.weapon.String()
	}
	a, err := game.PrepareStoredAction(ctx, testPool, c, pc, intent, false)
	f.must(err)
	return a.Preview
}
func (f *spatialFixture) engine() *engine.Engine {
	f.t.Helper()
	e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(f.dice))
	f.must(err)
	return e
}
func (f *spatialFixture) confirm(p *game.ActionPreview) (engine.PreparedAction, *engine.TurnResult) {
	f.t.Helper()
	if !p.Available {
		f.t.Fatalf("unavailable: %+v", p)
	}
	r := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
	if p.Intent.Kind == combat.EncounterStart || p.Intent.Kind == srd521.MartialAttack {
		r.TargetID = f.target
	}
	result, err := f.engine().ConfirmAction(context.Background(), f.campaign, r)
	f.must(err)
	return r, result
}

func TestIntegrationSpatialTargetsAndReload(t *testing.T) {
	f := newSpatialFixture(t)
	p := f.prepare(srd521.MartialAttack)
	if p.Available || len(p.Targets) != 1 || *p.Targets[0].DistanceFeet != 15 || !strings.Contains(p.Targets[0].UnavailableReason, "reach") {
		t.Fatalf("range preview: %+v", p)
	}
	f.confirm(f.prepare(combat.EncounterStart))
	for _, distance := range []int{10, 5} {
		f.confirm(f.prepare(combat.EncounterApproach))
		saved, err := game.ReadEncounter(context.Background(), testPool, f.campaign)
		f.must(err)
		if *saved.Space.DistanceFeet != distance {
			t.Fatalf("reload distance=%d", *saved.Space.DistanceFeet)
		}
	}
	f.confirm(f.prepare(srd521.MartialAttack))
	f.confirm(f.prepare(combat.EncounterRetreat))
	calls := f.dice.calls
	f.confirm(f.prepare(combat.EncounterEndTurn))
	if f.dice.calls != calls {
		t.Fatal("NPC attacked beyond five-foot reach")
	}
	if f.prepare(srd521.MartialAttack).Available {
		t.Fatal("out of range after turn/reload")
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"half"`)
	p = f.prepare(srd521.MartialAttack)
	if !p.Available {
		t.Fatalf("new ruling unavailable: %+v", p)
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"total"`)
	_, err := f.engine().ConfirmAction(context.Background(), f.campaign, engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent, TargetID: f.target})
	if !errors.Is(err, engine.ErrActionRefresh) || f.dice.calls != calls {
		t.Fatalf("stale cover rolled: %v", err)
	}
	p = f.prepare(srd521.MartialAttack)
	if p.Available || !strings.Contains(p.Targets[0].UnavailableReason, "Total cover") {
		t.Fatal("total cover admitted")
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"half"`)
	f.dice.roll = 6
	_, result := f.confirm(f.prepare(srd521.MartialAttack))
	var payload struct {
		Receipt srd521.AttackReceipt `json:"receipt"`
	}
	f.must(json.Unmarshal(result.AppliedToolCalls[0].Result, &payload))
	if payload.Receipt.Total != 12 || payload.Receipt.Outcome != "hit" {
		t.Fatalf("cover receipt: %+v", payload.Receipt)
	}
	f.confirm(f.prepare(combat.EncounterEndTurn))
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"three-quarters"`)
	_, result = f.confirm(f.prepare(srd521.MartialAttack))
	f.must(json.Unmarshal(result.AppliedToolCalls[0].Result, &payload))
	if payload.Receipt.Total != 12 || payload.Receipt.Outcome != "miss" {
		t.Fatalf("three-quarters receipt: %+v", payload.Receipt)
	}

}

func TestIntegrationSpatialClarificationAndPrivacy(t *testing.T) {
	f := newSpatialFixture(t)
	for _, facts := range []string{`"line_of_sight":true,"cover":"none"`, `"distance_feet":5,"cover":"none"`, `"distance_feet":5,"line_of_sight":true`} {
		f.facts(facts)
		p := f.prepare(srd521.MartialAttack)
		if p.Available || !strings.Contains(p.Targets[0].UnavailableReason, "Clarify") {
			t.Fatalf("unresolved ruling: %+v", p)
		}
	}
	f.facts(`"distance_feet":5,"line_of_sight":false,"cover":"none"`)
	if f.prepare(srd521.MartialAttack).Available {
		t.Fatal("unseen target admitted")
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"none"`)
	f.confirm(f.prepare(combat.EncounterStart))
	p := f.prepare(combat.EncounterEndTurn)
	encoded, err := json.Marshal(p)
	f.must(err)
	if strings.Contains(string(encoded), "unpublished villain") {
		t.Fatal("secret metadata leaked")
	}
	f.facts(`"distance_feet":5,"line_of_sight":true,"cover":"none","hidden":true`)
	p = f.prepare(combat.EncounterEndTurn)
	encoded, err = json.Marshal(p)
	f.must(err)
	if p.Available || p.Encounter != nil || strings.Contains(string(encoded), "Practice foe") {
		t.Fatalf("hidden actor leaked: %s", encoded)
	}
	if f.dice.calls != 2 {
		t.Fatal("clarification/preparation rolled")
	}
}
