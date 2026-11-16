//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

func TestIntegrationFaithRecipients(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Independent Faith", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	campaign := dbutil.FromPgtype(c.ID)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, campaign) })
	build, err := srd521.SupportedBuildForConcept("Human Cleric", "Acolyte")
	must(err)
	stats, err := build.Stats()
	must(err)
	actors := []uuid.UUID{}
	for _, name := range []string{"Mira", "Iona"} {
		pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: name, Stats: stats, Hp: 10, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
		must(err)
		actors = append(actors, dbutil.FromPgtype(pc.ID))
	}
	// This is deliberate test entitlement, never inferred from catalogue presence.
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,isolated_recipient_effects}','true')||'{"custom":{"keep":7}}'::jsonb WHERE campaign_id=$1`, campaign)
	must(err)
	var location uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name,location_type) VALUES($1,'Faith room','room') RETURNING id`, campaign).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$2 WHERE campaign_id=$1`, campaign, location)
	must(err)
	facts := map[string]any{"profile": "shield-of-faith-isolated-v1", "actor_ids": actors, "location_id": location, "hidden": false, "distance_feet": 60, "line_of_sight": true, "cover": "none"}
	raw, err := json.Marshal(facts)
	must(err)
	var target uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,alive,location_id,properties) VALUES($1,'Known guard',7,true,$2,jsonb_build_object('srd521_faith',$3::jsonb)) RETURNING id`, campaign, location, raw).Scan(&target))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved) VALUES($1,1,'fixture','game_action','known creature',$2)`, campaign, []uuid.UUID{target})
	must(err)
	dice := &lifecycleIntegrationDice{}
	newEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	prepare := func(actor uuid.UUID, kind string) *game.ActionPreview {
		t.Helper()
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		intent := game.ActionIntent{ActorID: actor, Kind: kind}
		if kind == srd521.SpellCast {
			intent.EntryID = "shield-of-faith"
			intent.SlotLevel = 1
		}
		a, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		return a.Preview
	}
	request := func(actor uuid.UUID, kind string) engine.PreparedAction {
		t.Helper()
		p := prepare(actor, kind)
		if !p.Available {
			t.Fatalf("unavailable: %+v", p)
		}
		r := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
		if kind == srd521.SpellCast {
			r.TargetID = target
			if len(p.Targets) != 2 {
				t.Fatalf("targets: %+v", p.Targets)
			}
		}
		return r
	}
	confirm := func(r engine.PreparedAction) *engine.TurnResult {
		t.Helper()
		v, err := newEngine().ConfirmAction(ctx, campaign, r)
		must(err)
		return v
	}
	active := func() *game.FaithEffect {
		t.Helper()
		now, err := game.ReadSpellClock(ctx, testPool, campaign)
		must(err)
		v, err := game.ReadFaithEffect(ctx, testPool, campaign, target, now)
		must(err)
		return v
	}
	// Exercise the actual authorized HTTP preparation/confirmation adapters.
	logger := log.New(io.Discard)
	ch := &handlers.CharacterHandlers{Queries: q, Pool: testPool, Logger: logger}
	ah := &handlers.ActionHandlers{Engine: newEngine(), Queries: q, Logger: logger}
	owner := &handlers.CampaignHandlers{Queries: q, Logger: logger}
	router := chi.NewRouter()
	router.Use(auth.NewNoOpMiddleware(testUserID).Authenticate)
	router.Route("/campaigns/{id}", func(r chi.Router) {
		r.Use(owner.RequireOwnership)
		r.Get("/character/action-preview", ch.PreviewAction)
		r.Post("/action/confirm", ah.ConfirmAction)
	})
	path := "/campaigns/" + campaign.String()
	for _, actor := range actors {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path+"/character/action-preview?actor_id="+actor.String()+"&kind=spell-cast&entry_id=shield-of-faith&slot_level=1", nil))
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), target.String()) || strings.Contains(rec.Body.String(), `"hp"`) {
			t.Fatalf("HTTP preparation: %d %s", rec.Code, rec.Body.String())
		}
	}
	for _, field := range []string{`"costs":[]`, `"started_at":0`, `"expires_at":999999`} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, path+"/action/confirm", strings.NewReader(`{"action_id":"`+uuid.NewString()+`","expected_revision":"forged","intent":{"actor_id":"`+actors[0].String()+`","kind":"spell-cast"},`+field+`}`)))
		if rec.Code != 400 {
			t.Fatalf("forged field accepted: %s %d", field, rec.Code)
		}
	}
	tuiPreview, err := newEngine().PreviewFaithAction(ctx, campaign, actors[0], false)
	must(err)
	if !tuiPreview.Available || len(tuiPreview.Targets) != 2 {
		t.Fatal("real TUI preparation disconnected")
	}
	before := request(actors[0], srd521.SpellCast)
	// Preview has neither charged slots nor rolled dice.
	if dice.calls != 0 {
		t.Fatal("preview rolled")
	}
	first := confirm(before)
	if active() == nil || active().ID != before.ActionID {
		t.Fatal("first identity missing")
	}
	replay := confirm(before)
	if !sameTurnResult(first, replay) {
		t.Fatal("fresh engine replay changed")
	}
	confirm(request(actors[0], srd521.WaitMinute))
	second := request(actors[1], srd521.SpellCast)
	confirm(second)
	if active().ID != second.ActionID {
		t.Fatal("latest equal cast not selected")
	}
	oldExpiry := active().ExpiresAt - 60
	confirm(request(actors[1], srd521.EndConcentration))
	if active().ID != before.ActionID || active().ExpiresAt != oldExpiry {
		t.Fatal("older cast failed to resume original expiry")
	}
	// Same caster replacement removes only its former link, retaining identities.
	replace := request(actors[0], srd521.SpellCast)
	replace.ActionID = uuid.MustParse("ffffffff-ffff-4fff-bfff-ffffffffffff")
	confirm(replace)
	if active().ID != replace.ActionID {
		t.Fatal("replacement failed")
	}
	var links, casts int
	must(testPool.QueryRow(ctx, `SELECT count(*) FROM concentration_recipients WHERE cast_id IN (SELECT id FROM concentration_casts WHERE campaign_id=$1)`, campaign).Scan(&links))
	must(testPool.QueryRow(ctx, `SELECT count(*) FROM concentration_casts WHERE campaign_id=$1`, campaign).Scan(&casts))
	if links != 1 || casts != 3 {
		t.Fatalf("links=%d casts=%d", links, casts)
	}
	// Recipient mutations invalidate the entire prepared snapshot.
	stale := request(actors[1], srd521.SpellCast)
	_, err = testPool.Exec(ctx, `UPDATE npcs SET hp=6 WHERE id=$1`, target)
	must(err)
	if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale recipient accepted: %v", err)
	}
	for _, change := range []string{`'hidden','true'::jsonb`, `'distance_feet','61'::jsonb`, `'line_of_sight','false'::jsonb`, `'cover','"total"'::jsonb`} {
		fields := strings.SplitN(change, ",", 2)
		_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_set(properties,ARRAY['srd521_faith',`+fields[0]+`],`+fields[1]+`) WHERE id=$1`, target)
		must(err)
		if len(prepare(actors[1], srd521.SpellCast).Targets) != 1 {
			t.Fatalf("invalid facts accepted: %s", change)
		}
		_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_build_object('srd521_faith',$2::jsonb) WHERE id=$1`, target, raw)
		must(err)
	}
	// Failure after all effect/cost mutations leaves no partial state or receipt.
	rollback := request(actors[1], srd521.SpellCast)
	rollback.ActionID = uuid.MustParse("00000000-0000-4000-8000-000000000001")
	snapshot := func() string {
		t.Helper()
		var value string
		must(testPool.QueryRow(ctx, `SELECT jsonb_build_object(
 'actors',(SELECT jsonb_agg(to_jsonb(p) ORDER BY id) FROM player_characters p WHERE campaign_id=$1),
 'clock',(SELECT jsonb_agg(to_jsonb(c)) FROM campaign_time c WHERE campaign_id=$1),
 'casts',(SELECT jsonb_agg(to_jsonb(c) ORDER BY id) FROM concentration_casts c WHERE campaign_id=$1),
 'links',(SELECT jsonb_agg(to_jsonb(r) ORDER BY cast_id) FROM concentration_recipients r WHERE cast_id IN (SELECT id FROM concentration_casts WHERE campaign_id=$1)),
 'history',(SELECT jsonb_agg(to_jsonb(l) ORDER BY id) FROM session_logs l WHERE campaign_id=$1),
 'receipts',(SELECT jsonb_agg(to_jsonb(a) ORDER BY id) FROM turn_attempts a WHERE campaign_id=$1))::text`, campaign).Scan(&value))
		return value
	}
	beforeFailure := snapshot()
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_faith_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.llm_response LIKE 'Shield of Faith on selected%' THEN RAISE EXCEPTION 'history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_faith_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_faith_history()`)
	must(err)
	_, failure := newEngine().ConfirmAction(ctx, campaign, rollback)
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_faith_history ON session_logs; DROP FUNCTION reject_faith_history()`)
	must(err)
	if snapshot() != beforeFailure {
		t.Fatal("history failure partially committed actors/costs/effects/links/clock/history/receipts")
	}
	if failure == nil {
		t.Fatal("history failure ignored")
	}
	var exists bool
	must(testPool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM concentration_casts WHERE id=$1)`, rollback.ActionID).Scan(&exists))
	if exists {
		t.Fatal("cast survived rollback")
	}
	confirm(rollback)
	// The later confirmation wins even with the lower UUID in the same second.
	var sameSecond bool
	must(testPool.QueryRow(ctx, `SELECT a.started_at=b.started_at FROM concentration_casts a,concentration_casts b WHERE a.id=$1 AND b.id=$2`, replace.ActionID, rollback.ActionID).Scan(&sameSecond))
	if !sameSecond {
		t.Fatal("ordering fixture did not share a campaign second")
	}
	var increasingOrder bool
	must(testPool.QueryRow(ctx, `SELECT a.cast_order<b.cast_order FROM concentration_casts a,concentration_casts b WHERE a.id=$1 AND b.id=$2`, replace.ActionID, rollback.ActionID).Scan(&increasingOrder))
	if !increasingOrder {
		t.Fatal("confirmation order was not persisted")
	}
	if active().ID != rollback.ActionID {
		t.Fatal("latest confirmation lost to UUID order")
	}
	c2, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actors[0], testUserID, false)
	must(err)
	_ = c2
	sheet, err := srd521.Project(p.Stats, 1, p.HP, p.MaxHP, 0)
	must(err)
	now, err := game.ReadSpellClock(ctx, testPool, campaign)
	must(err)
	must(game.ApplyFaithAC(ctx, testPool, campaign, actors[0], now, sheet))
	if sheet.Defenses["armor-class"].Value != 11 {
		t.Fatal("recipient spell granted caster AC")
	}
	// Derivation applies one bonus even across repeated projection/reload calls.
	recipientSheet, err := srd521.Project(stats, 1, 10, 10, 0)
	must(err)
	must(game.ApplyFaithAC(ctx, testPool, campaign, target, now, recipientSheet))
	must(game.ApplyFaithAC(ctx, testPool, campaign, target, now, recipientSheet))
	if recipientSheet.Defenses["armor-class"].Value != 13 {
		t.Fatal("overlapping/repeated projection doubled AC")
	}
	// Actor and effect edits independently invalidate prepared revisions.
	staleActor := request(actors[0], srd521.EndConcentration)
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=stats||'{"revision_test":true}'::jsonb WHERE id=$1`, actors[1])
	must(err)
	if _, err := newEngine().ConfirmAction(ctx, campaign, staleActor); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale actor accepted: %v", err)
	}
	staleEffect := request(actors[0], srd521.EndConcentration)
	_, err = testPool.Exec(ctx, `DELETE FROM concentration_recipients WHERE cast_id=$1`, rollback.ActionID)
	must(err)
	if _, err := newEngine().ConfirmAction(ctx, campaign, staleEffect); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale effect accepted: %v", err)
	}
	_, err = testPool.Exec(ctx, `INSERT INTO concentration_recipients(cast_id,recipient_id) VALUES($1,$2)`, rollback.ActionID, target)
	must(err)
	must(game.ApplyFaithAC(ctx, testPool, campaign, actors[0], now+600, sheet))
	if sheet.Concentration != nil || sheet.Defenses["armor-class"].Value != 11 {
		t.Fatal("public projection retained expired concentration")
	}
	// Concurrent independent engines serialize; exactly one stale revision wins.
	a := request(actors[0], srd521.EndConcentration)
	b := request(actors[0], srd521.EndConcentration)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, r := range []engine.PreparedAction{a, b} {
		wg.Add(1)
		go func(r engine.PreparedAction) {
			defer wg.Done()
			_, err := newEngine().ConfirmAction(ctx, campaign, r)
			errs <- err
		}(r)
	}
	wg.Wait()
	close(errs)
	successes := 0
	for err := range errs {
		if err == nil {
			successes++
		} else if !errors.Is(err, engine.ErrActionRefresh) && !errors.Is(err, engine.ErrTurnInProgress) {
			t.Fatal(err)
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent winners=%d", successes)
	}
	// Exact expiry is exclusive, never restarts or doubles.
	remaining := active()
	if remaining == nil {
		t.Fatal("other caster removed")
	}
	expired, err := game.ReadFaithEffect(ctx, testPool, campaign, target, remaining.ExpiresAt)
	must(err)
	if expired != nil {
		t.Fatal("exact expiry still active")
	}
	var kept bool
	must(testPool.QueryRow(ctx, `SELECT stats->'custom'='{"keep":7}'::jsonb FROM player_characters WHERE id=$1`, actors[0]).Scan(&kept))
	if !kept {
		t.Fatal("unknown JSON lost")
	}
	encoded, err := json.Marshal(first)
	must(err)
	if strings.Contains(string(encoded), `"hp":7`) || strings.Contains(string(encoded), "tactic") {
		t.Fatal("private recipient information disclosed")
	}
	// Independent recipients: self on one caster and known NPC on the other.
	confirm(request(actors[0], srd521.LongRest))
	confirm(request(actors[1], srd521.LongRest))
	self := request(actors[0], srd521.SpellCast)
	self.TargetID = actors[0]
	confirm(self)
	separate := request(actors[1], srd521.SpellCast)
	confirm(separate)
	now, err = game.ReadSpellClock(ctx, testPool, campaign)
	must(err)
	own, err := game.ReadFaithEffect(ctx, testPool, campaign, actors[0], now)
	must(err)
	if own == nil || own.ID != self.ActionID || active().ID != separate.ActionID {
		t.Fatal("independent targets overwritten")
	}
	confirm(request(actors[0], srd521.EndConcentration))
	if active().ID != separate.ActionID {
		t.Fatal("ending removed another caster's target link")
	}
	// Catalogue support without explicit isolated entitlement never discovers NPCs.
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,casting,isolated_recipient_effects}','false') WHERE id=$1`, actors[0])
	must(err)
	if len(prepare(actors[0], srd521.SpellCast).Targets) != 1 {
		t.Fatal("catalogue granted recipient entitlement")
	}

}
