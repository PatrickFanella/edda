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
	"github.com/charmbracelet/log"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/google/uuid"
)

func TestIntegrationAllySpellAction(t *testing.T) {
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
	facts := fmt.Sprintf(`{"profile":"cure-wounds-ally-v1","actor_id":"%s","location_id":"%s","allied":true,"willing":true,"conscious":true,"hidden":false,"conditions":"ordinary","max_hp":10,"distance_feet":5,"line_of_sight":true,"cover":"none"}`, actor, location)
	must(testPool.QueryRow(ctx, `INSERT INTO npcs(campaign_id,name,hp,location_id,properties) VALUES($1,'Rowan',1,$2,jsonb_build_object('srd521_ally',$3::jsonb,'secret','unpublished villain')) RETURNING id`, c.ID, location, facts).Scan(&ally))
	_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,npcs_involved) VALUES($1,1,'Arrive','game_action','Rowan waits.',ARRAY[$2]::uuid[])`, c.ID, ally)
	must(err)
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
	if !preview.Available || len(preview.Targets) != 2 || preview.Targets[0].ID != actor || preview.Costs[0].Amount != 1 {
		t.Fatalf("preview=%+v", preview)
	}
	dice := &castingDice{}
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
	path := "/character/action-preview?actor_id=" + actor.String() + "&kind=spell-cast&entry_id=cure-wounds&slot_level=1"
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
		must(testPool.QueryRow(ctx, `SELECT hp FROM npcs WHERE id=$1`, ally).Scan(&hp))
		must(testPool.QueryRow(ctx, `SELECT (stats->'srd521'->'casting'->>'slots')::int FROM player_characters WHERE id=$1`, pc.ID).Scan(&slots))
		must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
		return hp, slots, logs
	}
	if hp, slots, logs := read(); hp != 1 || slots != 2 || logs != 1 {
		t.Fatalf("partial write hp=%d slots=%d logs=%d", hp, slots, logs)
	}
	var attempts int
	must(testPool.QueryRow(ctx, `SELECT COUNT(*) FROM turn_attempts WHERE campaign_id=$1`, c.ID).Scan(&attempts))
	if attempts != 0 {
		t.Fatal("history failure left a receipt")
	}
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_cast_history ON session_logs; DROP FUNCTION reject_cast_history()`)
	must(err)
	body, err := json.Marshal(request)
	must(err)
	confirmed := call(testUserID, "POST", "/action/confirm", string(body))
	if confirmed.Code != 200 {
		t.Fatalf("HTTP confirm=%d %s", confirmed.Code, confirmed.Body.String())
	}
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	var casterHP int
	must(testPool.QueryRow(ctx, `SELECT hp FROM player_characters WHERE id=$1`, pc.ID).Scan(&casterHP))
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
	if hp, slots, logs := read(); hp != 8 || slots != 1 || logs != 2 || dice.calls != 4 {
		t.Fatalf("hp=%d slots=%d logs=%d dice=%d", hp, slots, logs, dice.calls)
	}
	var receipt struct {
		Receipt srd521.SpellReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.TargetID != ally.String() || receipt.Receipt.Target != "Rowan" || receipt.Receipt.Healing != 7 || receipt.Receipt.SlotsAfter != 1 {
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
		t.Fatal("empty slot still available")
	}
	saved, err := makeEngine().GetActionReceipt(ctx, campaign, first.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("reload receipt differs")
	}
}
