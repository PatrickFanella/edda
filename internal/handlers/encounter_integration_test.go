//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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

type encounterDice struct{ calls int }

func (d *encounterDice) Intn(n int) int {
	d.calls++
	if n == 20 {
		return 18
	}
	return 0
}
func TestIntegrationEncounterTurns(t *testing.T) {
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Encounter fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	var location, target, weapon, potion uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name)VALUES($1,'Court')RETURNING id`, c.ID).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, location, pc.ID)
	must(err)
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,properties)VALUES($1,$2,'Longsword','weapon','common',true,'{"srd521_id":"longsword"}')RETURNING id`, c.ID, pc.ID).Scan(&weapon))
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,quantity,properties)VALUES($1,$2,'Potion','consumable','common',3,'{"srd521_id":"potion-of-healing"}')RETURNING id`, c.ID, pc.ID).Scan(&potion))
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties)VALUES($1,'Practice foe',100,$2,$3)RETURNING id`, c.ID, location, fmt.Sprintf(`{"srd521_target":{"profile":"unarmored-fixture-v1","actor_id":"%s","distance_feet":5}}`, actor)).Scan(&target))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved)VALUES($1,1,'Arrive','game_action','Practice foe waits.',ARRAY[$2]::uuid[])`, c.ID, target)
	must(err)
	dice := &encounterDice{}
	newEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	prepare := func(kind, entry string) *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, game.ActionIntent{ActorID: actor, Kind: kind, EntryID: entry}, false)
		must(err)
		return a.Preview
	}
	confirm := func(p *game.ActionPreview, targetID uuid.UUID) (engine.PreparedAction, *engine.TurnResult) {
		t.Helper()
		if !p.Available {
			t.Fatalf("unavailable: %+v", p)
		}
		r := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent, TargetID: targetID}
		result, err := newEngine().ConfirmAction(ctx, campaign, r)
		must(err)
		return r, result
	}
	start := prepare(combat.EncounterStart, weapon.String())
	if dice.calls != 0 {
		t.Fatal("preview rolled")
	}
	request, result := confirm(start, target)
	id := request.ActionID
	encounter, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if encounter.Round != 1 || encounter.Active().ID != actor || !encounter.Budget.Action {
		t.Fatalf("start=%+v", encounter)
	}
	calls := dice.calls
	replay, err := newEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	if dice.calls != calls || !sameTurnResult(result, replay) {
		t.Fatal("start replay rolled")
	}
	attack := prepare(srd521.MartialAttack, weapon.String())
	confirm(attack, target)
	if prepare(srd521.MartialAttack, weapon.String()).Available {
		t.Fatal("double attack available")
	}
	stale := request
	stale.ActionID = uuid.New()
	if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale initiative: %v", err)
	}
	confirm(prepare(srd521.ConsumableUse, potion.String()), actor)
	if prepare(srd521.ConsumableUse, potion.String()).Available {
		t.Fatal("double bonus action available")
	}
	for i := 0; i < 6; i++ {
		confirm(prepare(combat.EncounterMove, ""), uuid.Nil)
	}
	if prepare(combat.EncounterMove, "").Available {
		t.Fatal("movement exceeded")
	}
	endRequest, endResult := confirm(prepare(combat.EncounterEndTurn, ""), uuid.Nil)
	encounter, err = game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if encounter.Round != 2 || encounter.Active().ID != actor || !encounter.Budget.Action || !encounter.Budget.BonusAction || encounter.Budget.Movement != 30 {
		t.Fatalf("reload=%+v", encounter)
	}
	var payload struct {
		NPCReceipts []srd521.AttackReceipt `json:"npc_receipts"`
	}
	must(json.Unmarshal(endResult.AppliedToolCalls[0].Result, &payload))
	if len(payload.NPCReceipts) != 1 || payload.NPCReceipts[0].Damage != 1 || payload.NPCReceipts[0].Die != 19 {
		t.Fatalf("NPC receipt=%+v", payload)
	}
	calls = dice.calls
	replay, err = newEngine().ConfirmAction(ctx, campaign, endRequest)
	must(err)
	if dice.calls != calls || !sameTurnResult(replay, endResult) {
		t.Fatal("NPC replay repeated")
	}
	var hp, quantity, logs int
	must(testPool.QueryRow(ctx, `SELECT hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&hp))
	must(testPool.QueryRow(ctx, `SELECT quantity FROM items WHERE id=$1`, potion).Scan(&quantity))
	must(testPool.QueryRow(ctx, `SELECT count(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
	if hp != 11 || quantity != 2 || logs != 11 {
		t.Fatalf("hp=%d quantity=%d logs=%d", hp, quantity, logs)
	}
	saved, err := newEngine().GetActionReceipt(ctx, campaign, id)
	must(err)
	if !sameTurnResult(saved.Result, result) {
		t.Fatal("receipt changed after later turns")
	}
	confirm(prepare(combat.EncounterDash, ""), uuid.Nil)
	encounter, err = game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if encounter.Budget.Action || encounter.Budget.Movement != 60 {
		t.Fatal("dash budget not saved")
	}
}
