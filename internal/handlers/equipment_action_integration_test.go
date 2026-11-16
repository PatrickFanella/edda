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

func TestIntegrationEquipmentAction(t *testing.T) {
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
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Equipment", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	must(err)
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	must(err)
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	var mail, shield uuid.UUID
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,quantity,properties) VALUES($1,$2,'Chain Mail','armor','common',true,1,'{"srd521_id":"chain-mail"}') RETURNING id`, c.ID, pc.ID).Scan(&mail))
	must(testPool.QueryRow(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,quantity,properties) VALUES($1,$2,'Shield','armor','common',true,1,'{"srd521_id":"shield"}') RETURNING id`, c.ID, pc.ID).Scan(&shield))
	intent := game.ActionIntent{ActorID: actor, Kind: srd521.EquipmentRemove, EntryID: mail.String()}

	h := &handlers.CharacterHandlers{Queries: q, Pool: testPool, Logger: log.New(io.Discard)}
	campaignH := &handlers.CampaignHandlers{Queries: q, Logger: log.New(io.Discard)}
	router := chi.NewRouter()
	router.Use(auth.NewNoOpMiddleware(testUserID).Authenticate)
	router.Route("/campaigns/{id}", func(r chi.Router) {
		r.Use(campaignH.RequireOwnership)
		r.Get("/character/action-preview", h.PreviewAction)
	})
	prepare := func() *game.ActionPreview {
		t.Helper()
		rec := httptest.NewRecorder()
		path := "/campaigns/" + campaign.String() + "/character/action-preview?actor_id=" + actor.String() + "&kind=" + intent.Kind + "&entry_id=" + intent.EntryID
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != 200 {
			t.Fatalf("preview status=%d body=%s", rec.Code, rec.Body.String())
		}
		var p game.ActionPreview
		must(json.Unmarshal(rec.Body.Bytes(), &p))
		return &p
	}

	dice := &potionDice{}
	makeEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		must(err)
		return e
	}
	if !prepare().Available {
		t.Fatalf("preview=%+v", prepare())
	}
	request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: prepare().ExpectedRevision, Intent: intent}
	// Ownership, depleted quantity, magic/custom properties and stale equipped state fail closed.
	for _, sql := range []string{`UPDATE items SET player_character_id=NULL WHERE id=$1`, `UPDATE items SET quantity=0 WHERE id=$1`, `UPDATE items SET properties='{"srd521_id":"chain-mail","magic_bonus":1}' WHERE id=$1`, `UPDATE items SET properties='{"srd521_id":"plate"}' WHERE id=$1`, `UPDATE items SET equipped=false WHERE id=$1`} {
		_, err = testPool.Exec(ctx, sql, mail)
		must(err)
		if prepare().Available {
			t.Fatal("negative fixture available")
		}
		if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("negative accepted: %v", err)
		}
		_, err = testPool.Exec(ctx, `UPDATE items SET player_character_id=$2,quantity=1,equipped=true,properties='{"srd521_id":"chain-mail"}' WHERE id=$1`, mail, pc.ID)
		must(err)
	}
	request.ExpectedRevision = prepare().ExpectedRevision
	_, err = testPool.Exec(ctx, `UPDATE items SET updated_at=clock_timestamp() WHERE id=$1`, shield)
	must(err)
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale equipped inventory accepted: %v", err)
	}
	request.ExpectedRevision = prepare().ExpectedRevision
	_, err = testPool.Exec(ctx, `CREATE FUNCTION reject_equipment_history() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.player_input='Confirm chain-mail equipment-remove' THEN RAISE EXCEPTION 'fixture history failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER reject_equipment_history BEFORE INSERT ON session_logs FOR EACH ROW EXECUTE FUNCTION reject_equipment_history()`)
	must(err)
	t.Cleanup(func() {
		_, _ = testPool.Exec(ctx, `DROP TRIGGER IF EXISTS reject_equipment_history ON session_logs; DROP FUNCTION IF EXISTS reject_equipment_history()`)
	})
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); err == nil {
		t.Fatal("history failure ignored")
	}
	assertState := func(wantEquipped bool, wantAC, wantClock, wantLogs int) {
		t.Helper()
		var equipped bool
		var saved []byte
		var logs, attempts int
		must(testPool.QueryRow(ctx, `SELECT equipped FROM items WHERE id=$1`, mail).Scan(&equipped))
		must(testPool.QueryRow(ctx, `SELECT stats FROM player_characters WHERE id=$1`, pc.ID).Scan(&saved))
		sheet, err := srd521.Project(saved, 1, 12, 12, 0)
		must(err)
		clock, err := game.ReadSpellClock(ctx, testPool, campaign)
		must(err)
		must(testPool.QueryRow(ctx, `SELECT count(*) FROM session_logs WHERE campaign_id=$1`, c.ID).Scan(&logs))
		must(testPool.QueryRow(ctx, `SELECT count(*) FROM turn_attempts WHERE campaign_id=$1`, c.ID).Scan(&attempts))
		if attempts != wantLogs {
			t.Fatalf("attempts=%d want=%d", attempts, wantLogs)
		}
		if equipped != wantEquipped || sheet.Defenses["armor-class"].Value != wantAC || clock != wantClock || logs != wantLogs {
			t.Fatalf("equipped=%t AC=%d clock=%d logs=%d", equipped, sheet.Defenses["armor-class"].Value, clock, logs)
		}
	}
	assertState(true, 19, 28800, 0)
	_, err = testPool.Exec(ctx, `DROP TRIGGER reject_equipment_history ON session_logs; DROP FUNCTION reject_equipment_history()`)
	must(err)
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	assertState(false, 13, 29100, 1)
	for i := 0; i < 3; i++ {
		replay, err := makeEngine().ConfirmAction(ctx, campaign, request)
		must(err)
		if !sameTurnResult(first, replay) {
			t.Fatal("replay changed")
		}
	}
	saved, err := makeEngine().GetActionReceipt(ctx, campaign, request.ActionID)
	must(err)
	if !sameTurnResult(first, saved.Result) {
		t.Fatal("reload changed")
	}
	assertState(false, 13, 29100, 1)
	intent.Kind = srd521.EquipmentEquip
	request = engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: prepare().ExpectedRevision, Intent: intent}
	_, err = makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	assertState(true, 19, 29700, 2)
	var receipt struct {
		Receipt srd521.EquipmentReceipt `json:"receipt"`
	}
	must(json.Unmarshal(first.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.ACBefore != 19 || receipt.Receipt.ACAfter != 13 || dice.calls != 0 {
		t.Fatalf("receipt=%+v dice=%d", receipt, dice.calls)
	}
	// Unknown/two-handed held equipment prevents a shield mutation.
	_, err = testPool.Exec(ctx, `INSERT INTO items(campaign_id,player_character_id,name,item_type,rarity,equipped,quantity,properties) VALUES($1,$2,'Greatsword','weapon','common',true,1,'{"srd521_id":"greatsword"}')`, c.ID, pc.ID)
	must(err)
	intent = game.ActionIntent{ActorID: actor, Kind: srd521.EquipmentRemove, EntryID: shield.String()}
	if prepare().Available {
		t.Fatal("unresolved hands accepted")
	}
	// Reload a caster fixture at an expiry boundary: one shield Utilize action
	// expires concentration atomically and receipt replay never advances twice.
	_, err = testPool.Exec(ctx, `DELETE FROM items WHERE campaign_id=$1 AND item_type='weapon'`, c.ID)
	must(err)
	_, err = testPool.Exec(ctx, `UPDATE items SET equipped=false WHERE id=$1`, mail)
	must(err)
	cleric, err := os.ReadFile("../rules/srd521/testdata/cleric1.json")
	must(err)
	_, err = testPool.Exec(ctx, `UPDATE player_characters SET hp=10,max_hp=10,stats=jsonb_set(jsonb_set($1::jsonb,'{srd521,shield}','true'),'{srd521,casting,concentration}','{"id":"expiry-fixture","spell":"shield-of-faith","started_at":29101,"expires_at":29701}') WHERE id=$2`, cleric, pc.ID)
	must(err)
	intent = game.ActionIntent{ActorID: actor, Kind: srd521.EquipmentRemove, EntryID: shield.String()}
	if p := prepare(); !p.Available {
		t.Fatalf("expiry preview: %+v", p)
	}
	request = engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: prepare().ExpectedRevision, Intent: intent}
	result, err := makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	must(json.Unmarshal(result.AppliedToolCalls[0].Result, &receipt))
	if receipt.Receipt.Expired == nil || receipt.Receipt.TimeAfter != 29706 || receipt.Receipt.ACAfter != 11 {
		t.Fatalf("expiry=%+v", receipt)
	}
	_, err = makeEngine().ConfirmAction(ctx, campaign, request)
	must(err)
	clock, err := game.ReadSpellClock(ctx, testPool, campaign)
	must(err)
	if clock != 29706 {
		t.Fatal("replay advanced clock")
	}
	var concentration []byte
	must(testPool.QueryRow(ctx, `SELECT stats#>'{srd521,casting,concentration}' FROM player_characters WHERE id=$1`, pc.ID).Scan(&concentration))
	if string(concentration) != "null" {
		t.Fatalf("effect retained: %s", concentration)
	}

}
