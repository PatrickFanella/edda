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

type castingDice struct{ calls int }

func (d *castingDice) Intn(int) int { d.calls++; return 1 }
func TestIntegrationSpellAction(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	stats, err := os.ReadFile("../rules/srd521/testdata/cleric1.json")
	must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Casting fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 1, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	intent := game.ActionIntent{ActorID: actor, Kind: srd521.SpellCast, EntryID: "cure-wounds", SlotLevel: 1}
	prepare := func() *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		return a.Preview
	}
	for _, field := range []string{"prepared", "verbal", "somatic"} {
		_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,ARRAY['srd521','casting',$1],'false') WHERE id=$2`, field, pc.ID)
		must(err)
		if prepare().Available {
			t.Fatalf("%s not enforced", field)
		}
		_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,ARRAY['srd521','casting',$1],'true') WHERE id=$2`, field, pc.ID)
		must(err)
	}
	intent.SlotLevel = 2
	if prepare().Available {
		t.Fatal("upcast accepted")
	}
	intent.SlotLevel = 1
	preview := prepare()
	if !preview.Available || len(preview.Targets) != 1 || preview.Targets[0].ID != actor || preview.Costs[0].Amount != 1 {
		t.Fatalf("preview=%+v", preview)
	}
	dice := &castingDice{}
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
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,slots}','1') WHERE id=$1`, pc.ID)
	must(err)
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale slot: %v", err)
	}
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,slots}','2') WHERE id=$1`, pc.ID)
	must(err)
	request.ExpectedRevision = prepare().ExpectedRevision
	if dice.calls != 0 {
		t.Fatal("preview/invalid action rolled")
	}
	// A history-write failure must roll back both HP and slots.
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_cast_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.player_input='Confirm Cure Wounds spell-cast' THEN RAISE EXCEPTION 'fixture history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_cast_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_cast_history()`)
	must(err)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_cast_history ON session_logs; DROP FUNCTION IF EXISTS reject_cast_history()`)
	})
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); err == nil {
		t.Fatal("history failure ignored")
	}
	read := func() (int, int, int) {
		var hp, slots, logs int
		must(testPool.QueryRow(ctx, `SELECT hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&hp))
		must(testPool.QueryRow(ctx, `SELECT (stats->'srd521'->'casting'->>'slots')::int FROM player_characters WHERE id=$1`, pc.ID).Scan(&slots))
		must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
		return hp, slots, logs
	}
	if hp, slots, logs := read(); hp != 1 || slots != 2 || logs != 0 {
		t.Fatalf("partial write hp=%d slots=%d logs=%d", hp, slots, logs)
	}
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_cast_history ON session_logs; DROP FUNCTION reject_cast_history()`)
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
	if hp, slots, logs := read(); hp != 8 || slots != 1 || logs != 1 || dice.calls != 4 {
		t.Fatalf("hp=%d slots=%d logs=%d dice=%d", hp, slots, logs, dice.calls)
	}
	var receipt struct {
		Receipt srd521.SpellReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.Healing != 7 || receipt.Receipt.SlotsAfter != 1 {
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
	if hp, slots, logs := read(); hp != 10 || slots != 0 || logs != 2 {
		t.Fatalf("capped hp=%d slots=%d logs=%d", hp, slots, logs)
	}
	if prepare().Available {
		t.Fatal("empty slot still available")
	}
	saved, err := makeEngine().GetActionReceipt(ctx, campaign, first.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("reload receipt differs")
	}
}
