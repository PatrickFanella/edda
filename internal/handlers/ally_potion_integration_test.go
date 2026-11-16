//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
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

func TestIntegrationAllyPotionAction(t *testing.T) {
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
	var location, ally uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name) VALUES($1,'Court') RETURNING id`, c.ID).Scan(&location))
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, location, pc.ID)
	must(err)
	facts := fmt.Sprintf(`{"profile":"potion-healing-ally-v1","actor_id":"%s","location_id":"%s","allied":true,"willing":true,"conscious":true,"hidden":false,"conditions":"ordinary","max_hp":10,"distance_feet":5,"line_of_sight":true,"cover":"none"}`, actor, location)
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties) VALUES($1,'Rowan',1,$2,jsonb_build_object('srd521_ally',$3::jsonb,'secret','unpublished villain')) RETURNING id`, c.ID, location, facts).Scan(&ally))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved) VALUES($1,1,'Arrive','game_action','Rowan waits.',ARRAY[$2]::uuid[])`, c.ID, ally)
	must(err)
	var item uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,quantity,properties) VALUES($1,$2,'Potion of Healing','consumable','common',2,'{"srd521_id":"potion-of-healing"}') RETURNING id`, c.ID, pc.ID).Scan(&item))
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
	if !preview.Available || len(preview.Targets) != 2 || preview.Targets[0].ID != actor || preview.Costs[0].Amount != 1 {
		t.Fatalf("preview=%+v", preview)
	}
	dice := &potionDice{}
	makeEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: intent, TargetID: ally}
	// Mutate real saved scene facts after preparation. Every case must fail
	// before dice, and a new preview must omit the ineligible ally entirely.
	for _, tc := range []struct{ name, sql string }{
		{"moved", `UPDATE npcs SET location_id=NULL WHERE id=$1`},
		{"hidden", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,hidden}','true') WHERE id=$1`},
		{"unwilling", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,willing}','false') WHERE id=$1`},
		{"unknown willingness", `UPDATE npcs SET properties=properties #- '{srd521_ally,willing}' WHERE id=$1`},
		{"not allied", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,allied}','false') WHERE id=$1`},
		{"unconscious", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,conscious}','false') WHERE id=$1`},
		{"zero hp", `UPDATE npcs SET hp=0 WHERE id=$1`},
		{"distant", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,distance_feet}','10') WHERE id=$1`},
		{"unresolved", `UPDATE npcs SET properties=properties #- '{srd521_ally,distance_feet}' WHERE id=$1`},
		{"blocked", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,cover}','"total"') WHERE id=$1`},
		{"unseen", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,line_of_sight}','false') WHERE id=$1`},
		{"condition", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,conditions}','"unconscious"') WHERE id=$1`},
		{"other actor", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,actor_id}','"00000000-0000-0000-0000-000000000000"') WHERE id=$1`},
		{"unknown conditions", `UPDATE npcs SET properties=properties #- '{srd521_ally,conditions}' WHERE id=$1`},
		{"unknown maximum", `UPDATE npcs SET properties=properties #- '{srd521_ally,max_hp}' WHERE id=$1`},
		{"unknown mechanic", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,healing_multiplier}','2') WHERE id=$1`},
		{"case alias", `UPDATE npcs SET properties=jsonb_set(properties #- '{srd521_ally,willing}','{srd521_ally,WILLING}','true') WHERE id=$1`},
		{"unsupported revision", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,revision}','"unsupported mechanic"') WHERE id=$1`},
		{"missing ruling location", `UPDATE npcs SET properties=properties #- '{srd521_ally,location_id}' WHERE id=$1`},
		{"wrong ruling location", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,location_id}','"00000000-0000-0000-0000-000000000001"') WHERE id=$1`},
		{"spell profile", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,profile}','"cure-wounds-ally-v1"') WHERE id=$1`},
		{"null consent", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,willing}','null') WHERE id=$1`},
		{"embedded fields", `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,Space}','{}') WHERE id=$1`},
		{"stale hp", `UPDATE npcs SET hp=2 WHERE id=$1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Cleanup(func() {
				_, err := testPool.Exec(ctx, `UPDATE npcs SET hp=1,location_id=$1,properties=jsonb_build_object('srd521_ally',$2::jsonb,'secret','unpublished villain') WHERE id=$3`, location, facts, ally)
				must(err)
			})
			request.ExpectedRevision = prepare().ExpectedRevision
			_, err = testPool.Exec(ctx, tc.sql, ally)
			must(err)
			if _, err = makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
				t.Fatalf("changed target accepted: %v", err)
			}
			if tc.name != "stale hp" && len(prepare().Targets) != 1 {
				t.Fatal("ineligible target disclosed")
			}
		})
	}
	t.Run("model relocation cannot reuse touch ruling", func(t *testing.T) {
		var other uuid.UUID
		must(testPool.QueryRow(ctx, `INSERT INTO locations(campaign_id,name) VALUES($1,'Other scene') RETURNING id`, c.ID).Scan(&other))
		moveNPC := tools.NewUpdateNPCHandler(game.NewNPCService(q))
		t.Cleanup(func() {
			_, err := testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, location, pc.ID)
			must(err)
			_, err = moveNPC.Handle(ctx, map[string]any{"npc_id": ally.String(), "location_id": location.String()})
			must(err)
			_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_build_object('srd521_ally',$1::jsonb) WHERE id=$2`, facts, ally)
			must(err)
		})
		_, err = testPool.Exec(ctx, `UPDATE player_characters SET current_location_id=$1 WHERE id=$2`, other, pc.ID)
		must(err)
		_, err = moveNPC.Handle(ctx, map[string]any{"npc_id": ally.String(), "location_id": other.String()})
		must(err)
		fresh := prepare()
		if len(fresh.Targets) != 1 {
			t.Fatal("old touch ruling offered in a new scene")
		}
		invalid := request
		invalid.ExpectedRevision = fresh.ExpectedRevision
		if _, err := makeEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("old touch ruling confirmed in a new scene: %v", err)
		}
		_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,location_id}',to_jsonb($1::text)) WHERE id=$2`, other.String(), ally)
		must(err)
		if len(prepare().Targets) != 2 {
			t.Fatal("fresh trusted touch ruling not offered in new scene")
		}
	})
	for _, tc := range []struct{ name, sql string }{
		{"foreign ownership", `UPDATE items SET player_character_id=NULL WHERE id=$1`},
		{"zero", `UPDATE items SET quantity=0 WHERE id=$1`},
		{"changed quantity", `UPDATE items SET quantity=3 WHERE id=$1`},
		{"stronger identity", `UPDATE items SET properties='{"srd521_id":"potion-of-healing-greater"}' WHERE id=$1`},
		{"unknown identity", `UPDATE items SET properties='{}' WHERE id=$1`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pending := request
			pending.ExpectedRevision = prepare().ExpectedRevision
			_, err := testPool.Exec(ctx, tc.sql, item)
			must(err)
			_, confirmErr := makeEngine().ConfirmAction(ctx, campaign, pending)
			_, err = testPool.Exec(ctx, `UPDATE items SET player_character_id=$1,quantity=2,properties='{"srd521_id":"potion-of-healing"}' WHERE id=$2`, actor, item)
			must(err)
			if !errors.Is(confirmErr, engine.ErrActionRefresh) {
				t.Fatalf("inventory change accepted: %v", confirmErr)
			}
		})
	}
	for _, key := range []string{"profile", "actor_id", "location_id", "allied", "willing", "conscious", "hidden", "conditions", "max_hp", "distance_feet", "line_of_sight", "cover"} {
		for _, nullValue := range []bool{false, true} {
			t.Run(fmt.Sprintf("mandatory %s null %t", key, nullValue), func(t *testing.T) {
				if nullValue {
					_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_set(properties,ARRAY['srd521_ally',$1],'null') WHERE id=$2`, key, ally)
				} else {
					_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=properties #- ARRAY['srd521_ally',$1] WHERE id=$2`, key, ally)
				}
				must(err)
				fresh := prepare()
				_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_build_object('srd521_ally',$1::jsonb) WHERE id=$2`, facts, ally)
				must(err)
				if len(fresh.Targets) != 1 {
					t.Fatal("incomplete facts disclosed ally")
				}
			})
		}
	}
	// Previously encountered is independent of physical presence and consent.
	_, err = testPool.Exec(ctx, `UPDATE session_logs SET npcs_involved='{}' WHERE campaign_id=$1`, c.ID)
	must(err)
	if len(prepare().Targets) != 1 {
		t.Fatal("unencountered ally disclosed")
	}
	if _, err = makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatal("unencountered target accepted")
	}
	_, err = testPool.Exec(ctx, `UPDATE session_logs SET npcs_involved=ARRAY[$1]::uuid[] WHERE campaign_id=$2`, ally, c.ID)
	must(err)
	t.Run("target revalidated by confirmation locking query", func(t *testing.T) {
		gate := &allyPotionQueryGate{entered: make(chan struct{}), release: make(chan struct{})}
		cfg := testPool.Config().Copy()
		cfg.ConnConfig.Tracer = gate
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		must(err)
		defer pool.Close()
		e, err := engine.New(pool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		pending := request
		pending.ExpectedRevision = prepare().ExpectedRevision
		done := make(chan error, 1)
		go func() { _, err := e.ConfirmAction(ctx, campaign, pending); done <- err }()
		<-gate.entered
		_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_set(properties,'{srd521_ally,willing}','false') WHERE id=$1`, ally)
		close(gate.release)
		must(err)
		if err := <-done; !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("current target accepted: %v", err)
		}
		_, err = testPool.Exec(ctx, `UPDATE npcs SET properties=jsonb_build_object('srd521_ally',$1::jsonb) WHERE id=$2`, facts, ally)
		must(err)
	})
	request.ExpectedRevision = prepare().ExpectedRevision
	// The authenticated HTTP adapter uses the same public preview/receipt. A
	// foreign owner cannot preview, confirm or reload this campaign's action.
	cfg := config.Config{}
	cfg.Server.JWTSecret = "ally-fixture"
	router := server.NewRouterWithProvider(log.New(io.Discard), makeEngine(), q, nil, testPool, testUserID, cfg, nil)
	call := func(user uuid.UUID, method, path, body string) *httptest.ResponseRecorder {
		token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
		must(err)
		req := httptest.NewRequest(method, "/api/v1/campaigns/"+campaign.String()+path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	path := "/character/action-preview?actor_id=" + actor.String() + "&kind=consumable-use&entry_id=" + item.String() + ""
	for _, endpoint := range []struct{ method, path string }{{"GET", path}, {"POST", "/action/confirm"}, {"GET", "/action?action_id=" + request.ActionID.String()}} {
		if rec := call(uuid.New(), endpoint.method, endpoint.path, `{}`); rec.Code != 404 {
			t.Fatalf("ownership status=%d", rec.Code)
		}
	}
	public := call(testUserID, "GET", path, "")
	if public.Code != 200 || !strings.Contains(public.Body.String(), "Rowan") || strings.Contains(public.Body.String(), "unpublished") || strings.Contains(public.Body.String(), "max_hp") {
		t.Fatalf("public projection=%s", public.Body.String())
	}
	var projected game.ActionPreview
	must(json.Unmarshal(public.Body.Bytes(), &projected))
	if projected.ExpectedRevision != prepare().ExpectedRevision {
		t.Fatal("shared preview differs")
	}
	invalid := request
	invalid.TargetID = uuid.New()
	if _, err := makeEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("target accepted: %v", err)
	}
	request.ExpectedRevision = prepare().ExpectedRevision
	if dice.calls != 0 {
		t.Fatal("preview/invalid action rolled")
	}
	// A history-write failure must roll back both HP and slots.
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_ally_potion_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.player_input='Confirm Potion of Healing consumable-use' THEN RAISE EXCEPTION 'fixture history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_ally_potion_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_ally_potion_history()`)
	must(err)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_ally_potion_history ON session_logs; DROP FUNCTION IF EXISTS reject_ally_potion_history()`)
	})
	beforeClock, err := game.ReadSpellClock(ctx, testPool, campaign)
	must(err)
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); err == nil {
		t.Fatal("history failure ignored")
	}
	read := func() (int, int, int) {
		var hp, slots, logs int
		must(testPool.QueryRow(ctx, `SELECT hp FROM npcs WHERE id=$1`, ally).Scan(&hp))
		must(testPool.QueryRow(ctx, `SELECT quantity FROM items WHERE id=$1`, item).Scan(&slots))
		must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
		return hp, slots, logs
	}
	if hp, slots, logs := read(); hp != 1 || slots != 2 || logs != 1 {
		t.Fatalf("partial write hp=%d slots=%d logs=%d", hp, slots, logs)
	}
	afterClock, err := game.ReadSpellClock(ctx, testPool, campaign)
	must(err)
	if beforeClock != afterClock {
		t.Fatal("history failure advanced clock")
	}
	var attempts int
	must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, c.ID).Scan(&attempts))
	if attempts != 0 {
		t.Fatal("history failure left a receipt")
	}
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_ally_potion_history ON session_logs; DROP FUNCTION reject_ally_potion_history()`)
	must(err)
	body, err := json.Marshal(request)
	must(err)
	confirmed := call(testUserID, "POST", "/action/confirm", string(body))
	if confirmed.Code != 200 {
		t.Fatalf("HTTP confirm=%d %s", confirmed.Code, confirmed.Body.String())
	}
	var httpResult api.TurnResult
	must(json.Unmarshal(confirmed.Body.Bytes(), &httpResult))
	if httpResult.ActionID != request.ActionID.String() || len(httpResult.ResolutionEvents) != 1 || httpResult.ResolutionEvents[0].Type != "srd_consumable" {
		t.Fatalf("HTTP receipt=%+v", httpResult)
	}
	publicReceipt, err := json.Marshal(httpResult.ResolutionEvents[0].Details["receipt"])
	must(err)
	var projectedHealing srd521.HealingReceipt
	must(json.Unmarshal(publicReceipt, &projectedHealing))
	if projectedHealing.TargetID != ally.String() || projectedHealing.Healing != 6 || projectedHealing.ItemID != item.String() || len(projectedHealing.Dice) != 2 {
		t.Fatalf("public receipt=%+v", projectedHealing)
	}
	// Reconnect through the actual shared WebSocket adapter after the HTTP
	// commit; its public result must agree without another dice/cost mutation.
	srv := httptest.NewServer(router)
	defer srv.Close()
	token, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
	must(err)
	socketCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	conn, _, err := websocket.Dial(socketCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/campaigns/"+campaign.String()+"/ws", &websocket.DialOptions{HTTPHeader: headers})
	must(err)
	defer conn.CloseNow()
	must(wsjson.Write(socketCtx, conn, map[string]any{"type": "confirm_action", "payload": request}))
	var envelope api.WebSocketMessageEnvelope
	must(wsjson.Read(socketCtx, conn, &envelope))
	var socketResult api.TurnResult
	must(json.Unmarshal(envelope.Payload, &socketResult))
	if envelope.Type != "result" || !reflect.DeepEqual(socketResult, httpResult) {
		t.Fatalf("socket/public result differs: %+v", envelope)
	}
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	var casterHP int
	must(testPool.QueryRow(ctx, `SELECT hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&casterHP))
	var slots int
	must(testPool.QueryRow(ctx, `SELECT (stats->'srd521'->'casting'->>'slots')::int FROM player_characters WHERE id=$1`, pc.ID).Scan(&slots))
	if slots != 2 {
		t.Fatal("potion spent spell slots")
	}
	if casterHP != 1 {
		t.Fatal("ally cast changed caster HP")
	}

	for i := 0; i < 3; i++ {
		replay, err := makeEngine().ConfirmAction(ctx, campaign, request)
		must(err)
		if !sameTurnResult(first, replay) {
			t.Fatal("replay differs")
		}
	}
	if hp, slots, logs := read(); hp != 7 || slots != 1 || logs != 2 || dice.calls != 4 {
		t.Fatalf("hp=%d slots=%d logs=%d dice=%d", hp, slots, logs, dice.calls)
	}
	var receipt struct {
		Receipt srd521.HealingReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.TargetID != ally.String() || receipt.Receipt.Target != "Rowan" || receipt.Receipt.Healing != 6 || receipt.Receipt.QuantityAfter != 1 {
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
	if hp, slots, logs := read(); hp != 10 || slots != 0 || logs != 3 {
		t.Fatalf("capped hp=%d slots=%d logs=%d", hp, slots, logs)
	}
	if prepare().Available {
		t.Fatal("empty inventory still available")
	}
	for _, sameID := range []bool{true, false} {
		t.Run(fmt.Sprintf("simultaneous last potion same ID %t", sameID), func(t *testing.T) {
			_, err := testPool.Exec(ctx, `UPDATE items SET quantity=1,updated_at=clock_timestamp() WHERE id=$1`, item)
			must(err)
			_, err = testPool.Exec(ctx, `UPDATE npcs SET hp=1,updated_at=clock_timestamp() WHERE id=$1`, ally)
			must(err)
			req := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: prepare().ExpectedRevision, Intent: intent, TargetID: ally}
			contender := req
			if !sameID {
				contender.ActionID = uuid.New()
			}
			gate := &allyPotionGateDice{entered: make(chan struct{}), release: make(chan struct{})}
			e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(gate))
			must(err)
			var beforeLogs, beforeAttempts int
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, campaign).Scan(&beforeLogs))
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, campaign).Scan(&beforeAttempts))
			done := make(chan error, 1)
			go func() { _, err := e.ConfirmAction(ctx, campaign, req); done <- err }()
			<-gate.entered
			_, secondErr := makeEngine().ConfirmAction(ctx, campaign, contender)
			close(gate.release)
			must(<-done)
			if !errors.Is(secondErr, engine.ErrTurnInProgress) {
				t.Fatalf("concurrent confirmation=%v", secondErr)
			}
			replay, retryErr := makeEngine().ConfirmAction(ctx, campaign, contender)
			if sameID {
				must(retryErr)
				if replay.ActionID != req.ActionID {
					t.Fatal("wrong replay")
				}
			} else if !errors.Is(retryErr, engine.ErrActionRefresh) {
				t.Fatalf("last potion spent twice: %v", retryErr)
			}
			for i := 0; i < 2; i++ {
				_, err := makeEngine().ConfirmAction(ctx, campaign, req)
				must(err)
			}
			hp, qty, logs := read()
			var attempts int
			must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, campaign).Scan(&attempts))
			if hp != 7 || qty != 0 || logs != beforeLogs+1 || attempts != beforeAttempts+1 || gate.calls != 2 {
				t.Fatalf("hp=%d qty=%d logs=%d attempts=%d dice=%d", hp, qty, logs, attempts, gate.calls)
			}
		})
	}

	saved, err := makeEngine().GetActionReceipt(ctx, campaign, first.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("reload receipt differs")
	}
	t.Run("two actor combat keeps self and omits ally", func(t *testing.T) {
		_, err := testPool.Exec(ctx, `UPDATE items SET quantity=1 WHERE id=$1`, item)
		must(err)
		var foe uuid.UUID
		props := fmt.Sprintf(`{"srd521_target":{"profile":"unarmored-fixture-v1","actor_id":"%s","distance_feet":5}}`, actor)
		must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties) VALUES($1,'Foe',10,$2,$3) RETURNING id`, campaign, location, props).Scan(&foe))
		enc, err := combat.NewEncounter(combat.EncounterActor{ID: actor, Name: "Mira", HP: 1, MaxHP: 10}, combat.EncounterActor{ID: foe, Name: "Foe", HP: 10, MaxHP: 10}, 3, &potionDice{})
		must(err)
		calls, err := json.Marshal([]any{map[string]any{"Result": map[string]any{"encounter": enc}}})
		must(err)
		_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved,tool_calls) SELECT $1,MAX(turn_number)+1,'Fight','game_action','Fight',ARRAY[$2]::uuid[],$3 FROM session_logs WHERE campaign_id=$1`, campaign, foe, calls)
		must(err)
		fresh := prepare()
		if !fresh.Available || len(fresh.Targets) != 1 || fresh.Targets[0].ID != actor {
			t.Fatalf("combat preview=%+v", fresh)
		}
		_, err = makeEngine().ConfirmAction(ctx, campaign, engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: fresh.ExpectedRevision, Intent: intent, TargetID: ally})
		if !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("combat ally accepted: %v", err)
		}
	})

}

// Channels pause a real transaction at its dice boundary, while campaign, actor,
// inventory and target locks are held. No scheduling sleeps are needed.
type allyPotionGateDice struct {
	entered, release chan struct{}
	once             sync.Once
	calls            int
}

func (d *allyPotionGateDice) Intn(int) int {
	d.once.Do(func() { close(d.entered); <-d.release })
	d.calls++
	return 1
}

type allyPotionQueryGate struct {
	entered, release chan struct{}
	once             sync.Once
}

func (g *allyPotionQueryGate) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if strings.Contains(data.SQL, "SELECT n.id,n.name,n.hp,n.updated_at") && strings.Contains(data.SQL, "FOR UPDATE OF n") {
		g.once.Do(func() { close(g.entered); <-g.release })
	}
	return ctx
}
func (*allyPotionQueryGate) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}
