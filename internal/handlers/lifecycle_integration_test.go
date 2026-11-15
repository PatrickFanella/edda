//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

type lifecycleIntegrationDice struct {
	values []int
	calls  int
}

func (d *lifecycleIntegrationDice) Intn(int) int {
	d.calls++
	if len(d.values) == 0 {
		return 0
	}
	v := d.values[0]
	d.values = d.values[1:]
	return v
}

func TestIntegrationCasterLifecycle(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	creation, err := srd521.SupportedBuildForConcept("Human Cleric", "Acolyte")
	must(err)
	stats, err := creation.Stats()
	must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Lifecycle fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 2, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=stats||'{"custom":"keep"}'::jsonb WHERE id=$1`, pc.ID)
	must(err)
	dice := &lifecycleIntegrationDice{}
	newEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	prepare := func(kind, entry string) *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		intent := game.ActionIntent{ActorID: actor, Kind: kind, EntryID: entry}
		if kind == srd521.SpellCast {
			intent.SlotLevel = 1
		}
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		return a.Preview
	}
	requestFor := func(p *game.ActionPreview) engine.PreparedAction {
		t.Helper()
		if !p.Available {
			t.Fatalf("unavailable: %+v", p)
		}
		r := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
		if p.Intent.Kind == srd521.SpellCast {
			r.TargetID = actor
		}
		return r
	}
	confirm := func(p *game.ActionPreview) (engine.PreparedAction, *engine.TurnResult) {
		t.Helper()
		r := requestFor(p)
		result, err := newEngine().ConfirmAction(ctx, campaign, r)
		must(err)
		return r, result
	}
	read := func() (*srd521.CharacterSheet, int) {
		t.Helper()
		p, err := q.GetPlayerCharacterByID(ctx, pc.ID)
		must(err)
		s, err := srd521.Project(p.Stats, 1, int(p.Hp), 10, 0)
		must(err)
		clock, err := game.ReadSpellClock(ctx, testPool, campaign)
		must(err)
		if !json.Valid(p.Stats) {
			t.Fatal("invalid stats")
		}
		return s, clock
	}
	// No preview rolls or writes, including the rest's future Hit Die.
	p := prepare(srd521.SpellCast, "shield-of-faith")
	prepare(srd521.ShortRest, "hit-die")
	if dice.calls != 0 {
		t.Fatal("preview rolled")
	}
	r, first := confirm(p)
	s, clock := read()
	if s.Concentration == nil || s.Concentration.ID != r.ActionID.String() || s.Concentration.ExpiresAt != clock+600 || s.Defenses["armor-class"].Value != 13 || s.Resources[0].Current != 1 {
		t.Fatalf("sheet=%+v clock=%d", s, clock)
	}
	for i := 0; i < 2; i++ {
		replay, err := newEngine().ConfirmAction(ctx, campaign, r)
		must(err)
		if !sameTurnResult(first, replay) {
			t.Fatal("cast replay changed")
		}
	}
	confirm(prepare(srd521.WaitMinute, ""))
	s, clock = read()
	if s.Concentration.ExpiresAt != 29400 || clock != 28860 {
		t.Fatalf("reload reset duration: %+v %d", s.Concentration, clock)
	}
	_, second := confirm(prepare(srd521.SpellCast, "shield-of-faith"))
	var replaced struct {
		Receipt srd521.EffectReceipt `json:"receipt"`
	}
	must(json.Unmarshal(second.AppliedToolCalls[0].Result, &replaced))
	if replaced.Receipt.Before.ID != r.ActionID.String() || replaced.Receipt.After.ID == replaced.Receipt.Before.ID {
		t.Fatal("replacement did not identify the old effect")
	}
	confirm(prepare(srd521.EndConcentration, ""))
	if prepare(srd521.EndConcentration, "").Available {
		t.Fatal("ended twice")
	}
	// A failed history write rolls back clock, recovery and resource writes.
	rest := requestFor(prepare(srd521.LongRest, ""))
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_lifecycle_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.player_input='Confirm rest-long rest-long' THEN RAISE EXCEPTION 'history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_lifecycle_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_lifecycle_history()`)
	must(err)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_lifecycle_history ON session_logs; DROP FUNCTION IF EXISTS reject_lifecycle_history()`)
	})
	if _, err := newEngine().ConfirmAction(ctx, campaign, rest); err == nil {
		t.Fatal("history failure ignored")
	}
	s, after := read()
	if after != clock || s.HP.Value != 2 || s.Resources[0].Current != 0 {
		t.Fatalf("partial rest: %+v %d", s, after)
	}
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_lifecycle_history ON session_logs; DROP FUNCTION reject_lifecycle_history()`)
	must(err)
	recovered, err := newEngine().ConfirmAction(ctx, campaign, rest)
	must(err)
	replay, err := newEngine().ConfirmAction(ctx, campaign, rest)
	must(err)
	if !sameTurnResult(recovered, replay) {
		t.Fatal("rest replay changed")
	}
	s, clock = read()
	if clock != 57660 || s.HP.Value != 10 || s.Resources[0].Current != 2 || s.Resources[1].Current != 1 || s.Resources[2].Current != 1 {
		t.Fatalf("recovery=%+v clock=%d", s, clock)
	}
	if prepare(srd521.LongRest, "").Available {
		t.Fatal("repeat Long Rest available")
	}
	// Quiet time expires the effect after exactly 600 campaign seconds.
	confirm(prepare(srd521.SpellCast, "shield-of-faith"))
	for i := 0; i < 9; i++ {
		confirm(prepare(srd521.WaitMinute, ""))
	}
	s, _ = read()
	if s.Concentration == nil {
		t.Fatal("expired early")
	}
	confirm(prepare(srd521.WaitMinute, ""))
	s, _ = read()
	if s.Concentration != nil || s.Defenses["armor-class"].Value != 11 {
		t.Fatal("expiry did not remove AC")
	}
	// Confirmed short rest spends one d8, recovers no slots, saves its roll,
	// and retains unrelated stats. A second use is unavailable.
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET hp=2 WHERE id=$1`, pc.ID)
	must(err)
	dice.values = []int{3}
	shortReq, short := confirm(prepare(srd521.ShortRest, "hit-die"))
	s, _ = read()
	if s.HP.Value != 8 || s.Resources[0].Current != 1 || s.Resources[1].Current != 0 || dice.calls != 1 {
		t.Fatalf("short rest=%+v calls=%d", s, dice.calls)
	}
	replay, err = newEngine().ConfirmAction(ctx, campaign, shortReq)
	must(err)
	if !sameTurnResult(short, replay) || dice.calls != 1 {
		t.Fatal("Hit Die replay rolled")
	}
	if prepare(srd521.ShortRest, "hit-die").Available {
		t.Fatal("missing Hit Die available")
	}
	var custom string
	must(testPool.QueryRow(ctx, `SELECT stats->>'custom' FROM player_characters WHERE id=$1`, pc.ID).Scan(&custom))
	if custom != "keep" {
		t.Fatal("unrelated stats lost")
	}
	// Explicit ordinary encounter: a newly created Cleric can defend without
	// inventing a weapon attack. Shield of Faith spends only the Bonus Action.
	var location, target uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name)VALUES($1,'Court')RETURNING id`, c.ID).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1,hp=10 WHERE id=$2`, location, pc.ID)
	must(err)
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties)VALUES($1,'Practice foe',100,$2,$3)RETURNING id`, c.ID, location, fmt.Sprintf(`{"srd521_target":{"profile":"unarmored-fixture-v1","actor_id":"%s","distance_feet":5}}`, actor)).Scan(&target))
	_, err = testPool.Exec(ctx, `UPDATE session_logs SET npcs_involved=ARRAY[$1]::uuid[] WHERE campaign_id=$2`, target, c.ID)
	must(err)
	start := requestFor(prepare(combat.EncounterStart, ""))
	start.TargetID = target
	dice.values = []int{18, 0}
	_, err = newEngine().ConfirmAction(ctx, campaign, start)
	must(err)
	if prepare(srd521.LongRest, "").Available || prepare(srd521.WaitMinute, "").Available {
		t.Fatal("downtime allowed in encounter")
	}
	// Keep a second slot available to distinguish the turn rule from exhaustion.
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,slots}','2') WHERE id=$1`, pc.ID)
	must(err)
	confirm(prepare(srd521.SpellCast, "shield-of-faith"))
	if prepare(srd521.SpellCast, "cure-wounds").Available {
		t.Fatal("two slotted spells allowed in one turn")
	}
	enc, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if !enc.Budget.Action || enc.Budget.BonusAction {
		t.Fatal("Shield of Faith spent wrong action")
	}
	if prepare(srd521.SpellCast, "shield-of-faith").Available {
		t.Fatal("double Bonus Action allowed")
	}
	// A hit followed by a failed save removes concentration atomically.
	dice.values = []int{18, 0}
	beforeCalls := dice.calls
	endReq, end := confirm(prepare(combat.EncounterEndTurn, ""))
	s, after = read()
	if s.Concentration != nil || s.HP.Value != 9 || after != clock+600+3600+6 || dice.calls != beforeCalls+2 {
		t.Fatalf("damage result=%+v clock=%d calls=%d", s, after, dice.calls)
	}
	var payload struct {
		Effects []srd521.EffectReceipt `json:"concentration_receipts"`
	}
	must(json.Unmarshal(end.AppliedToolCalls[0].Result, &payload))
	if len(payload.Effects) != 1 || payload.Effects[0].Check == nil || payload.Effects[0].Check.Outcome != "failure" {
		t.Fatalf("effects=%+v", payload)
	}
	replay, err = newEngine().ConfirmAction(ctx, campaign, endReq)
	must(err)
	if !sameTurnResult(end, replay) || dice.calls != beforeCalls+2 {
		t.Fatal("damage/save replay changed")
	}
	stale := endReq
	stale.ActionID = uuid.New()
	if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale action admitted: %v", err)
	}
	// Six-second rounds must expire the buff after 100 rounds, not after 100
	// individual actions and not after a rounded-up minute per round.
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,slots}','1') WHERE id=$1`, pc.ID)
	must(err)
	confirm(prepare(srd521.SpellCast, "shield-of-faith"))
	s, startClock := read()
	expiry := s.Concentration.ExpiresAt
	for i := 0; i < 99; i++ {
		confirm(prepare(combat.EncounterEndTurn, ""))
	}
	s, nearExpiry := read()
	if s.Concentration == nil || nearExpiry != startClock+594 {
		t.Fatalf("combat expiry was early: %+v clock=%d", s.Concentration, nearExpiry)
	}
	_, expired := confirm(prepare(combat.EncounterEndTurn, ""))
	s, after = read()
	if s.Concentration != nil || after != expiry || s.Defenses["armor-class"].Value != 11 {
		t.Fatalf("combat expiry missing: %+v clock=%d", s.Concentration, after)
	}
	must(json.Unmarshal(expired.AppliedToolCalls[0].Result, &payload))
	if len(payload.Effects) != 1 || payload.Effects[0].Kind != "concentration-expired" {
		t.Fatalf("expiry receipt=%+v", payload)
	}

}
