//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

type martialDice struct{ calls int }

func (d *martialDice) Intn(n int) int {
	d.calls++
	if n == 20 {
		return 11
	}
	return 4
}
func TestIntegrationMartialAction(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Martial fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	var location, target, weapon, hidden uuid.UUID
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name)VALUES($1,'Courtyard')RETURNING id`, c.ID).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, location, pc.ID)
	must(err)
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,properties)VALUES($1,$2,'Longsword','weapon','common',true,'{"srd521_id":"longsword"}')RETURNING id`, c.ID, pc.ID).Scan(&weapon))
	props := fmt.Sprintf(`{"srd521_target":{"profile":"unarmored-fixture-v1","actor_id":"%s","distance_feet":5}}`, actor)
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties)VALUES($1,'Practice foe',20,$2,$3)RETURNING id`, c.ID, location, props).Scan(&target))
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties)VALUES($1,'Hidden foe',20,$2,$3)RETURNING id`, c.ID, location, props).Scan(&hidden))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved)VALUES($1,1,'Arrive','game_action','Practice foe waits.',ARRAY[$2]::uuid[])`, c.ID, target)
	must(err)
	intent := game.ActionIntent{ActorID: actor, Kind: srd521.MartialAttack, EntryID: weapon.String()}
	prepare := func() *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		return a.Preview
	}
	preview := prepare()
	if !preview.Available || len(preview.Targets) != 1 || preview.Targets[0].ID != target || preview.Modifier.Value != 5 {
		t.Fatalf("preview=%+v", preview)
	}
	dice := &martialDice{}
	makeEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: intent, TargetID: target}
	for _, id := range []uuid.UUID{hidden, uuid.New(), uuid.Nil} {
		invalid := request
		invalid.TargetID = id
		if _, err := makeEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("target admitted: %v", err)
		}
	}
	_, err = testPool.Exec(ctx, `UPDATE items SET equipped=false WHERE id=$1`, weapon)
	must(err)
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale weapon: %v", err)
	}
	_, err = testPool.Exec(ctx, `UPDATE items SET equipped=true WHERE id=$1`, weapon)
	must(err)
	request.ExpectedRevision = prepare().ExpectedRevision
	if dice.calls != 0 {
		t.Fatal("preview/invalid action rolled")
	}
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	replay, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	if !sameTurnResult(first, replay) || dice.calls != 2 {
		t.Fatal("retry rerolled")
	}
	var receipt struct {
		Receipt srd521.AttackReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.Damage != 8 || receipt.Receipt.Total != 17 {
		t.Fatalf("receipt=%+v", receipt)
	}
	var hp, logs int
	must(testPool.QueryRow(ctx, `SELECT hp FROM npcs WHERE id=$1`, target).Scan(&hp))
	must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
	if hp != 12 || logs != 2 {
		t.Fatalf("hp=%d logs=%d", hp, logs)
	}
	stale := request
	stale.ActionID = uuid.New()
	if _, err := makeEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale action: %v", err)
	}
	saved, err := makeEngine().GetActionReceipt(ctx, campaign, request.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("receipt reload differs")
	}
}
