//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	"github.com/charmbracelet/log"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/combat"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

func TestIntegrationActionSurge(t *testing.T) {
	ctx := context.Background()
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	q := statedb.New(testPool)
	creation, err := srd521.SupportedBuildForConcept("Human Fighter", "Soldier")
	must(err)
	other, err := q.CreateUser(ctx, "surge-other-owner")
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM users WHERE id=$1`, other.ID) })
	stats, err := creation.Stats()
	must(err)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Encounter fixture", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 20, Level: 2, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET experience=300,stats=jsonb_set(stats,'{srd521,advancement}','{"choice":"fixed","die":6,"hp_gain":8}'::jsonb) WHERE id=$1`, pc.ID)
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

	if prepare(srd521.ActionSurge, "").Available {
		t.Fatal("surge outside encounter")
	}
	confirm(prepare(combat.EncounterStart, weapon.String()), target)
	before, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	p := prepare(srd521.ActionSurge, "")
	cfg := config.Config{}
	cfg.Server.JWTSecret = "surge-test"
	router := server.NewRouterWithProvider(log.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
	body, err := json.Marshal(engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent})
	must(err)
	token, err := auth.GenerateToken(uuid.New(), cfg.Server.JWTSecret, time.Hour)
	must(err)
	req := httptest.NewRequest("POST", "/api/v1/campaigns/"+campaign.String()+"/action/confirm", strings.NewReader(string(body)))
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Fatalf("other owner %d", rec.Code)
	}

	after, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if !p.Available || !reflect.DeepEqual(before, after) || dice.calls != 2 {
		t.Fatal("preview mutated")
	}

	for _, change := range []string{"actor", "target", "revision", "intent"} {
		invalid := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
		switch change {
		case "actor":
			invalid.Intent.ActorID = uuid.New()
		case "target":
			invalid.TargetID = target
		case "revision":
			invalid.ExpectedRevision = "stale"
		case "intent":
			invalid.Intent.EntryID = "unsupported"
		}
		if _, err := newEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("invalid %s: %v", change, err)
		}
	}
	// The TUI engine method must use exactly the same preview contract.
	tuiPreview, err := newEngine().PreviewActionSurge(ctx, campaign)
	must(err)
	if tuiPreview.ExpectedRevision != p.ExpectedRevision || !tuiPreview.Available {
		t.Fatal("TUI preview differs")
	}
	// Changed actor, encounter identity and cursor invalidate the original preview.
	for _, change := range []string{"hp", "ownership", "encounter", "turn"} {
		stale := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
		if change == "ownership" {
			var attemptsBefore, historyBefore int
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, c.ID).Scan(&attemptsBefore))
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&historyBefore))
			_, err = testPool.Exec(ctx, `UPDATE player_characters SET user_id=$2 WHERE id=$1`, pc.ID, other.ID)
			must(err)
			if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
				t.Fatalf("stale ownership: %v", err)
			}
			// Refreshing must not authorize spending another user's resource.
			ownerToken, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
			must(err)
			previewPath := "/api/v1/campaigns/" + campaign.String() + "/character/action-preview?actor_id=" + actor.String() + "&kind=action-surge"
			freshRequest := httptest.NewRequest("GET", previewPath, nil)
			freshRequest.Header.Set("Authorization", "Bearer "+ownerToken)
			freshResponse := httptest.NewRecorder()
			router.ServeHTTP(freshResponse, freshRequest)
			if freshResponse.Code != 404 {
				t.Fatalf("reassigned character must not disclose a fresh preview: %d", freshResponse.Code)
			}
			// No new revision can be obtained; forged reuse remains rejected.
			fresh := p
			freshBody, err := json.Marshal(engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: fresh.ExpectedRevision, Intent: fresh.Intent})
			must(err)
			freshRequest = httptest.NewRequest("POST", "/api/v1/campaigns/"+campaign.String()+"/action/confirm", strings.NewReader(string(freshBody)))
			freshRequest.Header.Set("Authorization", "Bearer "+ownerToken)
			freshResponse = httptest.NewRecorder()
			router.ServeHTTP(freshResponse, freshRequest)
			if freshResponse.Code != 409 {
				t.Fatalf("fresh ownership confirmation status=%d", freshResponse.Code)
			}
			var spent bool
			must(testPool.QueryRow(ctx, `SELECT COALESCE((stats->'srd521'->>'action_surge_spent')::bool,false) FROM player_characters WHERE id=$1`, pc.ID).Scan(&spent))
			unchanged, err := game.ReadEncounter(ctx, testPool, campaign)
			must(err)
			var attemptsAfter, historyAfter int
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, c.ID).Scan(&attemptsAfter))
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&historyAfter))
			if spent || !reflect.DeepEqual(unchanged, before) || dice.calls != 2 || attemptsAfter != attemptsBefore || historyAfter != historyBefore {
				t.Fatal("ownership rejection changed resource, encounter, dice or receipt")
			}
			_, err = testPool.Exec(ctx, `UPDATE player_characters SET user_id=$1 WHERE id=$2`, testUserID, pc.ID)
			must(err)
			continue
		}
		if change == "hp" {
			_, err = testPool.Exec(ctx, `UPDATE player_characters SET hp=11 WHERE id=$1`, pc.ID)
			must(err)
		} else {
			var savedCalls []byte
			must(testPool.QueryRow(ctx, `SELECT tool_calls FROM session_logs WHERE campaign_id=$1 ORDER BY turn_number DESC LIMIT 1`, c.ID).Scan(&savedCalls))
			mutated := *before
			if change == "encounter" {
				mutated.ID = uuid.New()
			} else {
				mutated.Turn = 1
			}
			payload, err := json.Marshal([]map[string]any{{"Tool": "srd_encounter", "Result": map[string]any{"encounter": mutated}}})
			must(err)
			_, err = testPool.Exec(ctx, `UPDATE session_logs SET tool_calls=$1 WHERE campaign_id=$2 AND turn_number=(SELECT MAX(turn_number) FROM session_logs WHERE campaign_id=$2)`, payload, c.ID)
			must(err)
			if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
				t.Fatalf("stale %s: %v", change, err)
			}
			_, err = testPool.Exec(ctx, `UPDATE session_logs SET tool_calls=$1 WHERE campaign_id=$2 AND turn_number=(SELECT MAX(turn_number) FROM session_logs WHERE campaign_id=$2)`, savedCalls, c.ID)
			must(err)
			continue
		}
		if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("stale actor: %v", err)
		}
		_, err = testPool.Exec(ctx, `UPDATE player_characters SET hp=12 WHERE id=$1`, pc.ID)
		must(err)
	}
	stale := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,action_surge_spent}','true'::jsonb) WHERE id=$1`, pc.ID)
	must(err)
	if _, err := newEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale resource %v", err)
	}
	if prepare(srd521.ActionSurge, "").Available {
		t.Fatal("zero resource")
	}
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=stats #- '{srd521,action_surge_spent}' WHERE id=$1`, pc.ID)
	must(err)
	p = prepare(srd521.ActionSurge, "")
	request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
	_, err = testPool.Exec(ctx, `ALTER TABLE session_logs ADD CONSTRAINT surge_failure CHECK(player_input NOT LIKE 'Confirm Action Surge%')`)
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `ALTER TABLE session_logs DROP CONSTRAINT IF EXISTS surge_failure`) })
	if _, err := newEngine().ConfirmAction(ctx, campaign, request); err == nil {
		t.Fatal("expected history failure")
	}
	after, err = game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if !reflect.DeepEqual(before, after) || !prepare(srd521.ActionSurge, "").Available {
		t.Fatal("rollback lost resource or budget")
	}
	_, err = testPool.Exec(ctx, `ALTER TABLE session_logs DROP CONSTRAINT surge_failure`)
	must(err)
	first, err := newEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	replay, err := newEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	if !sameTurnResult(first, replay) || dice.calls != 2 {
		t.Fatal("fresh engine replay mutated")
	}
	saved, err := newEngine().GetActionReceipt(ctx, campaign, request.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("receipt reload differs")
	}
	after, err = game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if !after.Budget.Action || !after.Budget.SurgeAction || !after.Budget.SurgeUsed || !after.Budget.BonusAction || after.Budget.Movement != 30 || !after.Reaction.Available || after.Round != 1 {
		t.Fatalf("budget %+v", after)
	}
	if prepare(srd521.ActionSurge, "").Available {
		t.Fatal("second surge")
	}
	for _, intent := range []game.ActionIntent{{ActorID: actor, Kind: "magic"}, {ActorID: actor, Kind: "unsupported"}, {ActorID: actor, Kind: srd521.SpellCast, EntryID: "cure-wounds", SlotLevel: 1}, {ActorID: actor, Kind: srd521.ActionSurge, EntryID: "unsupported"}} {
		c, p, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		must(err)
		invalid, err := game.PrepareStoredAction(ctx, testPool, c, p, intent, false)
		must(err)
		if invalid.Preview.Available {
			t.Fatalf("invalid available %+v", intent)
		}
	}
	confirm(prepare(srd521.MartialAttack, weapon.String()), target)
	mid, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if mid.Budget.Action || !mid.Budget.SurgeAction || mid.Round != 1 {
		t.Fatal("first attack lost extra")
	}
	confirm(prepare(srd521.MartialAttack, weapon.String()), target)
	if prepare(srd521.MartialAttack, weapon.String()).Available {
		t.Fatal("third attack")
	}
	confirm(prepare(combat.EncounterEndTurn, ""), uuid.Nil)
	if prepare(srd521.ActionSurge, "").Available {
		t.Fatal("turn reset recovered resource")
	}
	// Spend ordinary first on the next turn, then recover only outside combat.
	confirm(prepare(srd521.MartialAttack, weapon.String()), target)
	_, err = testPool.Exec(ctx, `UPDATE npcs SET hp=0,alive=false WHERE id=$1`, target)
	must(err)
	// Persist an explicitly completed encounter as adjudicated fixture setup.
	enc, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	enc.Status = "completed"
	raw, err := json.Marshal(map[string]any{"encounter": enc})
	must(err)
	calls, err := json.Marshal([]map[string]any{{"Tool": "srd_encounter", "Result": json.RawMessage(raw)}})
	must(err)
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) SELECT $1,MAX(turn_number)+1,'complete','game_action','complete',$2 FROM session_logs WHERE campaign_id=$1`, c.ID, calls)
	must(err)
	confirm(prepare(srd521.ShortRest, ""), uuid.Nil)
	pcAfter, err := q.GetPlayerCharacterByID(ctx, pc.ID)
	must(err)
	var envelope struct {
		Build srd521.Build `json:"srd521"`
	}
	must(json.Unmarshal(pcAfter.Stats, &envelope))
	if srd521.ActionSurgeUses(&envelope.Build) != 1 {
		t.Fatal("short rest did not recover")
	}

	_, err = testPool.Exec(ctx, `UPDATE npcs SET hp=100,alive=true WHERE id=$1`, target)
	must(err)
	confirm(prepare(combat.EncounterStart, weapon.String()), target)
	confirm(prepare(srd521.MartialAttack, weapon.String()), target)
	confirm(prepare(srd521.ActionSurge, ""), uuid.Nil)
	remaining, err := game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	if remaining.Budget.Action || !remaining.Budget.SurgeAction {
		t.Fatal("surge after ordinary lost grant")
	}
	confirm(prepare(srd521.MartialAttack, weapon.String()), target)
	if prepare(srd521.MartialAttack, weapon.String()).Available {
		t.Fatal("third action after late surge")
	}
	// Rest is unavailable while a legitimate encounter is active.
	if prepare(srd521.ShortRest, "").Available || prepare(srd521.LongRest, "").Available {
		t.Fatal("rest during encounter")
	}
	enc, err = game.ReadEncounter(ctx, testPool, campaign)
	must(err)
	enc.Status = "completed"
	raw, err = json.Marshal(map[string]any{"encounter": enc})
	must(err)
	calls, err = json.Marshal([]map[string]any{{"Tool": "srd_encounter", "Result": json.RawMessage(raw)}})
	must(err)
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) SELECT $1,MAX(turn_number)+1,'complete','game_action','complete',$2 FROM session_logs WHERE campaign_id=$1`, c.ID, calls)
	must(err)
	confirm(prepare(srd521.LongRest, ""), uuid.Nil)
	if prepare(srd521.LongRest, "").Available {
		t.Fatal("early long rest")
	}
}
