//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
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

type potionDice struct{ calls int }

func (d *potionDice) Intn(int) int { d.calls++; return 1 }
func TestIntegrationConsumableAction(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Potion fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 1, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	var item uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,quantity,properties)VALUES($1,$2,'Potion of Healing','consumable','common',2,'{"srd521_id":"potion-of-healing"}')RETURNING id`, c.ID, pc.ID).Scan(&item))
	intent := game.ActionIntent{ActorID: actor, Kind: srd521.ConsumableUse, EntryID: item.String()}
	prepare := func() *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		return a.Preview
	}
	preview := prepare()
	if !preview.Available || len(preview.Targets) != 1 || preview.Targets[0].ID != actor || preview.Costs[0].Amount != 1 {
		t.Fatalf("preview=%+v", preview)
	}
	dice := &potionDice{}
	makeEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: intent, TargetID: actor}
	invalid := request
	invalid.TargetID = uuid.New()
	if _, err := makeEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("target accepted: %v", err)
	}
	_, err = testPool.Exec(ctx, `UPDATE items SET quantity=1 WHERE id=$1`, item)
	must(err)
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale quantity: %v", err)
	}
	_, err = testPool.Exec(ctx, `UPDATE items SET quantity=2 WHERE id=$1`, item)
	must(err)
	request.ExpectedRevision = prepare().ExpectedRevision
	if dice.calls != 0 {
		t.Fatal("preview/invalid action rolled")
	}
	// A history-write failure must roll back both HP and quantity.
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_potion_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.player_input='Confirm Potion of Healing consumable-use' THEN RAISE EXCEPTION 'fixture history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_potion_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_potion_history()`)
	must(err)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_potion_history ON session_logs; DROP FUNCTION IF EXISTS reject_potion_history()`)
	})
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); err == nil {
		t.Fatal("history failure ignored")
	}
	read := func() (int, int, int) {
		var hp, qty, logs int
		must(testPool.QueryRow(ctx, `SELECT hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&hp))
		must(testPool.QueryRow(ctx, `SELECT quantity FROM items WHERE id=$1`, item).Scan(&qty))
		must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
		return hp, qty, logs
	}
	if hp, qty, logs := read(); hp != 1 || qty != 2 || logs != 0 {
		t.Fatalf("partial write hp=%d qty=%d logs=%d", hp, qty, logs)
	}
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_potion_history ON session_logs; DROP FUNCTION reject_potion_history()`)
	must(err)
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	for i := 0; i < 3; i++ {
		replay, err := makeEngine().ConfirmAction(ctx, campaign, request)
		must(err)
		if !sameTurnResult(first, replay) {
			t.Fatal("replay differs")
		}
	}
	if hp, qty, logs := read(); hp != 7 || qty != 1 || logs != 1 || dice.calls != 4 {
		t.Fatalf("hp=%d qty=%d logs=%d dice=%d", hp, qty, logs, dice.calls)
	}
	var receipt struct {
		Receipt srd521.HealingReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.Healing != 6 || receipt.Receipt.QuantityAfter != 1 {
		t.Fatalf("receipt=%+v", receipt)
	}
	stale := request
	stale.ActionID = uuid.New()
	if _, err := makeEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale receipt accepted: %v", err)
	}
	request.ActionID = uuid.New()
	request.ExpectedRevision = prepare().ExpectedRevision
	_, err = makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	if hp, qty, logs := read(); hp != 12 || qty != 0 || logs != 2 {
		t.Fatalf("capped hp=%d qty=%d logs=%d", hp, qty, logs)
	}
	if prepare().Available {
		t.Fatal("empty potion still available")
	}
	saved, err := makeEngine().GetActionReceipt(ctx, campaign, first.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("reload receipt differs")
	}
}
