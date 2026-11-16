//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type easyAthleticsDice struct {
	mu       sync.Mutex
	d20, d10 int
	calls    []int
}

func (d *easyAthleticsDice) Intn(n int) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.calls = append(d.calls, n)
	switch n {
	case 20:
		return d.d20
	case 10:
		return d.d10
	default:
		panic("unexpected fixture die")
	}
}

func TestIntegrationEasyAthleticsProductPath(t *testing.T) {
	for _, scenario := range []string{"use", "still-failed", "decline", "base-success", "ordinary-unknown", "base-history-rollback", "window-rollback", "history-edit", "history-replace", "origin", "background", "forged"} {
		t.Run(scenario, func(t *testing.T) {
			ctx := context.Background()
			q := statedb.New(testPool)
			must := func(err error) {
				t.Helper()
				if err != nil {
					t.Fatal(err)
				}
			}
			creation, err := srd521.SupportedBuildForConcept("Human Fighter", "Soldier")
			must(err)
			stats, err := creation.Stats()
			must(err)
			var envelope map[string]json.RawMessage
			must(json.Unmarshal(stats, &envelope))
			var build srd521.Build
			must(json.Unmarshal(envelope["srd521"], &build))
			sheet, err := srd521.Project(stats, 1, 9, 12, 300)
			must(err)
			_, err = srd521.AdvanceFighter(&build, sheet, "fixed", nil)
			must(err)
			envelope["srd521"], err = json.Marshal(build)
			must(err)
			envelope["custom"] = json.RawMessage(`{"keep":true}`)
			stats, err = json.Marshal(envelope)
			must(err)
			c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Easy Athletics product path", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
			must(err)
			t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
			pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 20, Level: 2, Experience: 300, Status: "active", Abilities: []byte(`[]`)})
			must(err)
			campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
			dice := &easyAthleticsDice{d20: 3, d10: 0}
			if scenario == "still-failed" {
				dice.d20 = 0
			}
			if scenario == "base-success" {
				dice.d20 = 4
			}
			newEngine := func() *engine.Engine {
				e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
				must(err)
				return e
			}
			cfg := config.Config{}
			cfg.Server.JWTSecret = "easy-athletics-product-fixture"
			router := server.NewRouterWithProvider(log.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
			token, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
			must(err)
			basePath := "/api/v1/campaigns/" + campaign.String()
			request := func(method, path, body string) *httptest.ResponseRecorder {
				t.Helper()
				r := httptest.NewRequest(method, basePath+path, strings.NewReader(body))
				r.Header.Set("Authorization", "Bearer "+token)
				r.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				return w
			}
			preview := func(intent game.ActionIntent) *game.ActionPreview {
				t.Helper()
				v := url.Values{"actor_id": {actor.String()}, "kind": {intent.Kind}}
				if intent.EntryID != "" {
					v.Set("entry_id", intent.EntryID)
				}
				if intent.Ability != "" {
					v.Set("ability", intent.Ability)
				}
				if intent.Skill != "" {
					v.Set("skill", intent.Skill)
				}
				if intent.SlotLevel != 0 {
					v.Set("slot_level", fmt.Sprint(intent.SlotLevel))
				}
				w := request("GET", "/character/action-preview?"+v.Encode(), "")
				if w.Code != 200 {
					t.Fatalf("preview %d %s", w.Code, w.Body.String())
				}
				var p game.ActionPreview
				must(json.Unmarshal(w.Body.Bytes(), &p))
				return &p
			}
			confirm := func(r engine.PreparedAction, code int) *httptest.ResponseRecorder {
				t.Helper()
				data, err := json.Marshal(r)
				must(err)
				w := request("POST", "/action/confirm", string(data))
				if w.Code != code {
					t.Fatalf("confirm expected %d got %d %s", code, w.Code, w.Body.String())
				}
				return w
			}
			snapshot := func() string {
				t.Helper()
				var s string
				must(testPool.QueryRow(ctx, `SELECT jsonb_build_object('campaign',to_jsonb(c),'pc',(SELECT to_jsonb(p) FROM player_characters p WHERE p.id=$2),'logs',(SELECT jsonb_agg(l ORDER BY l.id) FROM session_logs l WHERE l.campaign_id=c.id),'attempts',(SELECT jsonb_agg(a ORDER BY a.id) FROM turn_attempts a WHERE a.campaign_id=c.id),'windows',(SELECT jsonb_agg(w ORDER BY w.actor_id) FROM tactical_mind_windows w WHERE w.campaign_id=c.id))::text FROM campaigns c WHERE c.id=$1`, c.ID, pc.ID).Scan(&s))
				return s
			}
			entry := game.ActionIntent{ActorID: actor, Kind: srd521.AbilityCheck, EntryID: game.EasyAthleticsEntry}
			if scenario == "ordinary-unknown" {
				entry.EntryID = ""
				entry.Skill = "athletics"
			}
			if scenario == "origin" || scenario == "background" {
				field, value := "origin", "Elf"
				if scenario == "background" {
					field, value = "background", "Acolyte"
				}
				_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,ARRAY['srd521_creation',$1],to_jsonb($2::text)) WHERE id=$3`, field, value, pc.ID)
				must(err)
			}
			before := snapshot()
			p := preview(entry)
			if snapshot() != before || len(dice.calls) != 0 {
				t.Fatal("preview mutated state or rolled")
			}
			if scenario == "origin" || scenario == "background" {
				if p.Available {
					t.Fatal("unsupported creation offered fixture")
				}
				return
			}
			if !p.Available {
				t.Fatalf("fixture unavailable %+v", p)
			}
			tui, err := newEngine().PreviewEasyAthletics(ctx, campaign)
			must(err)
			if scenario != "ordinary-unknown" && tui.ExpectedRevision != p.ExpectedRevision {
				t.Fatal("TUI preview differs")
			}
			baseRequest := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: p.ExpectedRevision, Intent: p.Intent}
			if scenario == "forged" {
				for _, bad := range []game.ActionIntent{{ActorID: actor, Kind: srd521.AbilityCheck, EntryID: "easy-strength-athletics-dc11"}, {ActorID: actor, Kind: srd521.SavingThrow, EntryID: game.EasyAthleticsEntry}, {ActorID: actor, Kind: srd521.MartialAttack, EntryID: game.EasyAthleticsEntry}, {ActorID: actor, Kind: srd521.AbilityCheck, EntryID: game.EasyAthleticsEntry, Ability: "strength"}, {ActorID: actor, Kind: srd521.AbilityCheck, EntryID: game.EasyAthleticsEntry, Skill: "athletics"}, {ActorID: actor, Kind: srd521.AbilityCheck, EntryID: game.EasyAthleticsEntry, SlotLevel: 1}} {
					if preview(bad).Available {
						t.Fatalf("forged preview %+v", bad)
					}
					forged := baseRequest
					forged.ActionID = uuid.New()
					forged.Intent = bad
					confirm(forged, 409)
				}
				for _, field := range []string{`"difficulty":1`, `"outcome":"failure"`, `"dc":1`} {
					raw, _ := json.Marshal(baseRequest)
					body := strings.TrimSuffix(string(raw), "}") + "," + field + "}"
					if w := request("POST", "/action/confirm", body); w.Code != 400 {
						t.Fatalf("forged field %s %d", field, w.Code)
					}
				}
				if snapshot() != before || len(dice.calls) != 0 {
					t.Fatal("forged action mutated")
				}
				return
			}
			if scenario == "base-history-rollback" || scenario == "window-rollback" {
				table, constraint, expression := "session_logs", "easy_history_failure", "player_input NOT LIKE 'Confirm Easy Strength%'"
				if scenario == "window-rollback" {
					table, constraint, expression = "tactical_mind_windows", "easy_window_failure", "difficulty<>10"
				}
				_, err = testPool.Exec(ctx, "ALTER TABLE "+table+" ADD CONSTRAINT "+constraint+" CHECK ("+expression+")")
				must(err)
				t.Cleanup(func() { _, _ = testPool.Exec(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT IF EXISTS "+constraint) })
				confirm(baseRequest, 500)
				if snapshot() != before {
					t.Fatal("base history/window/receipt partial commit")
				}
				_, err = testPool.Exec(ctx, "ALTER TABLE "+table+" DROP CONSTRAINT "+constraint)
				must(err)
				dice.calls = nil
			}
			first := confirm(baseRequest, 200)
			var baseResult struct {
				Narrative string `json:"narrative"`
				Events    []struct {
					Details struct {
						Receipt srd521.CheckReceipt `json:"receipt"`
					} `json:"details"`
				} `json:"resolution_events"`
			}
			must(json.Unmarshal(first.Body.Bytes(), &baseResult))
			if len(baseResult.Events) != 1 {
				t.Fatal(first.Body.String())
			}
			base := baseResult.Events[0].Details.Receipt
			outcome := "failure"
			if scenario == "base-success" {
				outcome = "success"
			}
			if scenario == "ordinary-unknown" {
				outcome = "unknown"
			}
			if base.Outcome != outcome || len(base.Dice) != 1 || base.Dice[0].Value != dice.d20+1 || (scenario != "ordinary-unknown" && (base.DC == nil || *base.DC != 10)) || (scenario == "ordinary-unknown" && base.DC != nil) {
				t.Fatalf("base %+v", base)
			}
			if !reflect.DeepEqual(dice.calls, []int{20}) {
				t.Fatal("base rolled d10", dice.calls)
			}
			// Replay through a fresh engine/router uses the durable receipt, not dice.
			router = server.NewRouterWithProvider(log.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
			replay := confirm(baseRequest, 200)
			var a, b any
			must(json.Unmarshal(first.Body.Bytes(), &a))
			must(json.Unmarshal(replay.Body.Bytes(), &b))
			if !reflect.DeepEqual(a, b) || len(dice.calls) != 1 {
				t.Fatal("base replay changed")
			}
			reloaded, err := q.GetPlayerCharacterByID(ctx, pc.ID)
			must(err)
			var e struct {
				Build srd521.Build `json:"srd521"`
			}
			must(json.Unmarshal(reloaded.Stats, &e))
			if e.Build.SecondWind != 2 || reloaded.Hp != 9 {
				t.Fatal("base spent or healed")
			}
			choice := "use"
			if scenario == "decline" {
				choice = "decline"
			}
			tactical := preview(game.ActionIntent{ActorID: actor, Kind: srd521.TacticalMind, EntryID: choice})
			if scenario == "base-success" || scenario == "ordinary-unknown" {
				if tactical.Available {
					t.Fatal("window opened without known failure")
				}
				return
			}
			if !tactical.Available {
				t.Fatalf("unreachable tactical preview %+v", tactical)
			}
			tacticalRequest := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: tactical.ExpectedRevision, Intent: tactical.Intent}
			if scenario == "history-edit" || scenario == "history-replace" {
				if scenario == "history-edit" {
					_, err = testPool.Exec(ctx, `UPDATE session_logs SET llm_response='Changed base' WHERE campaign_id=$1`, c.ID)
				} else {
					_, err = testPool.Exec(ctx, `UPDATE session_logs SET id=gen_random_uuid() WHERE campaign_id=$1`, c.ID)
				}
				must(err)
				if preview(tactical.Intent).Available {
					t.Fatal("altered original history accepted")
				}
				confirm(tacticalRequest, 409)
				if len(dice.calls) != 1 {
					t.Fatal("altered history rolled")
				}
				return
			}
			before = snapshot()
			if !preview(tactical.Intent).Available || snapshot() != before || len(dice.calls) != 1 {
				t.Fatal("tactical preview mutated")
			}
			result := confirm(tacticalRequest, 200)
			var tacticalResult struct {
				Events []struct {
					Details struct {
						Receipt srd521.TacticalMindReceipt `json:"receipt"`
					} `json:"details"`
				} `json:"resolution_events"`
			}
			must(json.Unmarshal(result.Body.Bytes(), &tacticalResult))
			if len(tacticalResult.Events) != 1 {
				t.Fatal(result.Body.String())
			}
			r := tacticalResult.Events[0].Details.Receipt
			wantDie, wantUses, wantOutcome := 1, 1, "success"
			if scenario == "decline" {
				wantDie, wantUses, wantOutcome = 0, 2, "declined"
			}
			if scenario == "still-failed" {
				wantUses, wantOutcome = 2, "failure"
			}
			if r.Die != wantDie || r.SecondWindBefore != 2 || r.SecondWindAfter != wantUses || r.Outcome != wantOutcome || r.CheckID != baseRequest.ActionID.String() || !reflect.DeepEqual(r.Base, base) || r.Total != base.Total+wantDie {
				t.Fatalf("tactical %+v base %+v", r, base)
			}
			before = snapshot()
			router = server.NewRouterWithProvider(log.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
			confirm(tacticalRequest, 200)
			confirm(baseRequest, 200)
			if snapshot() != before {
				t.Fatal("replay mutated database")
			}
			wantDice := []int{20, 10}
			if scenario == "decline" {
				wantDice = []int{20}
			}
			if !reflect.DeepEqual(dice.calls, wantDice) {
				t.Fatal("replay rerolled", dice.calls)
			}
			var original string
			must(testPool.QueryRow(ctx, `SELECT llm_response FROM session_logs WHERE campaign_id=$1 AND turn_number=1`, c.ID).Scan(&original))
			if original != baseResult.Narrative {
				t.Fatal("original narration changed")
			}
			tacticalRequest.ActionID = uuid.New()
			confirm(tacticalRequest, 409)
		})
	}
}
