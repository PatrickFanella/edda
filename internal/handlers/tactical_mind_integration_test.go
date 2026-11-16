//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	clog "github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type tacticalDice struct {
	mu           sync.Mutex
	value, calls int
}

func (d *tacticalDice) Intn(n int) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	if n != 10 {
		panic("base check rerolled")
	}
	d.calls++
	return d.value
}

func TestIntegrationTacticalMind(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, choice := range []string{"success", "failure", "decline", "ownership", "stale", "concurrent", "history-stale", "empty-resource", "origin-create", "background-create", "origin-consume", "background-consume", "history-edit", "history-replace", "history-missing-create"} {
		t.Run(choice, func(t *testing.T) {
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
			c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Tactical Mind", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
			must(err)
			t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
			p, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 20, Level: 2, Experience: 300, Status: "active", Abilities: []byte(`[]`)})
			must(err)
			campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(p.ID)
			// Explicit trusted adjudication: recorded base d20=4, Athletics total=9,
			// known failure, hidden DC. The original narration is already durable.
			_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,1,'base check','game_action','Original failed-check narration','[]')`, c.ID)
			must(err)
			dc := 10
			if choice == "failure" {
				dc = 20
			}
			if choice == "origin-create" || choice == "background-create" {
				field, value := "origin", "Elf"
				if choice == "background-create" {
					field, value = "background", "Sage"
				}
				_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,ARRAY['srd521_creation',$2],to_jsonb($3::text)) WHERE id=$1`, p.ID, field, value)
				must(err)
			}
			if choice == "history-missing-create" {
				_, err = testPool.Exec(ctx, `DELETE FROM session_logs WHERE campaign_id=$1`, c.ID)
				must(err)
			}
			ca, pc, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
			must(err)
			sheet, err = srd521.Project(stats, 2, 9, 20, 300)
			must(err)
			resolution, err := srd521.ResolveCheck(sheet, srd521.Check{Kind: srd521.AbilityCheck, Ability: "strength", Skill: "athletics"}, srd521.Difficulty{Value: dc, OutcomeKnown: true}, &lifecycleIntegrationDice{values: []int{3}})
			must(err)
			tx, err := testPool.Begin(ctx)
			must(err)
			err = game.OpenTacticalMindWindow(ctx, tx, ca, pc, uuid.New(), resolution, dc)
			if strings.HasSuffix(choice, "-create") {
				must(tx.Rollback(ctx))
				if err == nil {
					t.Fatal("unsupported creation or missing original history opened a window")
				}
				return
			}
			must(err)
			must(tx.Commit(ctx))
			entry := "use"
			if choice == "decline" {
				entry = "decline"
			}
			intent := game.ActionIntent{ActorID: actor, Kind: srd521.TacticalMind, EntryID: entry}
			prepare := func() *game.ActionPreview {
				t.Helper()
				ca, pc, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, uuid.Nil, false)
				must(err)
				a, err := game.PrepareStoredAction(ctx, testPool, ca, pc, intent, false)
				must(err)
				return a.Preview
			}
			preview := prepare()
			if !preview.Available {
				t.Fatalf("%+v", preview)
			}
			data, _ := json.Marshal(preview)
			if strings.Contains(string(data), `"dc"`) || strings.Contains(string(data), "difficulty") {
				t.Fatal(string(data))
			}
			dice := &tacticalDice{value: 0}
			newEngine := func() *engine.Engine {
				e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
				must(err)
				return e
			}
			tuiPreview, err := newEngine().PreviewTacticalMind(ctx, campaign, entry)
			must(err)
			if !tuiPreview.Available || tuiPreview.ExpectedRevision != preview.ExpectedRevision {
				t.Fatal("TUI preparation diverged")
			}
			request := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: preview.ExpectedRevision, Intent: preview.Intent}
			if choice == "origin-consume" || choice == "background-consume" || choice == "history-edit" || choice == "history-replace" {
				switch choice {
				case "origin-consume", "background-consume":
					field, value := "origin", "Elf"
					if choice == "background-consume" {
						field, value = "background", "Sage"
					}
					_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,ARRAY['srd521_creation',$2],to_jsonb($3::text)) WHERE id=$1`, p.ID, field, value)
					must(err)
					// A stored revision alone cannot qualify an unsupported build.
					ca, pc, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
					must(err)
					basePreview, err := game.PrepareCharacterAction(ctx, ca, pc, game.ActionIntent{ActorID: actor, Kind: srd521.AbilityCheck, Ability: "strength"})
					must(err)
					_, err = testPool.Exec(ctx, `UPDATE tactical_mind_windows SET revision=$2 WHERE actor_id=$1`, p.ID, basePreview.ExpectedRevision)
					must(err)
				case "history-edit":
					_, err = testPool.Exec(ctx, `UPDATE session_logs SET llm_response='Original check withdrawn' WHERE campaign_id=$1 AND turn_number=1`, c.ID)
					must(err)
				case "history-replace":
					_, err = testPool.Exec(ctx, `DELETE FROM session_logs WHERE campaign_id=$1 AND turn_number=1`, c.ID)
					must(err)
					_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,1,'base check','game_action','Original failed-check narration','[]')`, c.ID)
					must(err)
				}
				if prepare().Available {
					t.Fatal("unsupported build or edited original history offered Tactical Mind")
				}
				if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
					t.Fatal(err)
				}
				var uses int
				var resolved bool
				must(testPool.QueryRow(ctx, `SELECT (p.stats#>>'{srd521,second_wind}')::int,w.resolved FROM player_characters p JOIN tactical_mind_windows w ON w.actor_id=p.id WHERE p.id=$1`, p.ID).Scan(&uses, &resolved))
				if dice.calls != 0 || uses != 2 || resolved {
					t.Fatal("rejection changed dice, resource or window")
				}
				return
			}
			if choice == "ownership" {
				// Reassign the character while leaving the campaign owned by the caller.
				other := uuid.New()
				_, err = testPool.Exec(ctx, `INSERT INTO users(id,name,email,password_hash) VALUES($1,'Other',$2,'fixture')`, other, other.String()+"@fixture.invalid")
				must(err)
				t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM users WHERE id=$1`, other) })
				_, err = testPool.Exec(ctx, `UPDATE player_characters SET user_id=$1 WHERE id=$2`, other, p.ID)
				must(err)
				h := &handlers.CharacterHandlers{Queries: q, Pool: testPool, Logger: clog.Default()}
				router := chi.NewRouter()
				router.Get("/campaigns/{id}/character/action-preview", h.PreviewAction)
				r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/campaigns/%s/character/action-preview?actor_id=%s&kind=tactical-mind&entry_id=use", campaign, actor), nil)
				r = r.WithContext(auth.WithUser(r.Context(), testUserID))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, r)
				if w.Code != 404 {
					t.Fatalf("fresh authenticated preview: %d %s", w.Code, w.Body.String())
				}
				if prepare().Available {
					t.Fatal("fresh trusted preview ignored reassignment")
				}
				if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
					t.Fatal(err)
				}
				// Even reassigning the campaign as well cannot transfer the old window.
				_, err = testPool.Exec(ctx, `UPDATE campaigns SET created_by=$1 WHERE id=$2`, other, c.ID)
				must(err)
				r = r.WithContext(auth.WithUser(r.Context(), other))
				w = httptest.NewRecorder()
				router.ServeHTTP(w, r)
				if w.Code != 200 || strings.Contains(w.Body.String(), `"available":true`) {
					t.Fatalf("new owner fresh preview: %d %s", w.Code, w.Body.String())
				}
				if dice.calls != 0 {
					t.Fatal("ownership rolled")
				}
				return
			}
			if choice == "history-stale" {
				_, err = testPool.Exec(ctx, `INSERT INTO session_logs(campaign_id,turn_number,player_input,input_type,llm_response,tool_calls) VALUES($1,2,'later action','game_action','Later narration','[]')`, c.ID)
				must(err)
				if prepare().Available {
					t.Fatal("window survived later history")
				}
				if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
					t.Fatal(err)
				}
				if dice.calls != 0 {
					t.Fatal("stale window rolled")
				}
				return
			}
			if choice == "empty-resource" {
				_, err = testPool.Exec(ctx, `UPDATE player_characters SET stats=jsonb_set(stats,'{srd521,second_wind}','0') WHERE id=$1`, p.ID)
				must(err)
				if prepare().Available {
					t.Fatal("empty resource offered")
				}
				if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
					t.Fatal(err)
				}
				if dice.calls != 0 {
					t.Fatal("spent resource rolled")
				}
				return
			}
			if choice == "stale" {
				_, err = testPool.Exec(ctx, `UPDATE player_characters SET hp=8 WHERE id=$1`, p.ID)
				must(err)
				if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
					t.Fatal(err)
				}
				if dice.calls != 0 {
					t.Fatal("stale rolled")
				}
				return
			}
			// Real history failure must roll back both window closure and resource debit.
			_, err = testPool.Exec(ctx, `ALTER TABLE session_logs ADD CONSTRAINT tactical_history_failure CHECK(player_input NOT LIKE 'Confirm Tactical Mind%')`)
			must(err)
			_, err = newEngine().ConfirmAction(ctx, campaign, request)
			if err == nil {
				t.Fatal("history failure missing")
			}
			_, err = testPool.Exec(ctx, `ALTER TABLE session_logs DROP CONSTRAINT tactical_history_failure`)
			must(err)
			after, err := q.GetPlayerCharacterByID(ctx, p.ID)
			must(err)
			if string(after.Stats) == "" || after.Hp != 9 || !prepare().Available {
				t.Fatal("partial rollback")
			}
			failedCalls := dice.calls
			if choice == "concurrent" {
				start := make(chan struct{})
				var wg sync.WaitGroup
				for range 2 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						<-start
						_, err := newEngine().ConfirmAction(ctx, campaign, request)
						if err != nil && !errors.Is(err, engine.ErrTurnInProgress) {
							t.Error(err)
						}
					}()
				}
				close(start)
				wg.Wait()
			}
			first, err := newEngine().ConfirmAction(ctx, campaign, request)
			must(err)
			if choice == "concurrent" {
				var wg sync.WaitGroup
				for range 2 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, err := newEngine().ConfirmAction(ctx, campaign, request)
						if err != nil && !errors.Is(err, engine.ErrTurnInProgress) {
							t.Error(err)
						}
					}()
				}
				wg.Wait()
			}
			cfg := config.Config{}
			router := server.NewRouterWithProvider(clog.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
			body, err := json.Marshal(request)
			must(err)
			httpRequest := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+campaign.String()+"/action/confirm", strings.NewReader(string(body)))
			httpResponse := httptest.NewRecorder()
			router.ServeHTTP(httpResponse, httpRequest)
			if httpResponse.Code != 200 || !strings.Contains(httpResponse.Body.String(), `"srd_tactical_mind"`) || strings.Contains(httpResponse.Body.String(), `"dc"`) {
				t.Fatalf("HTTP durable replay: %d %s", httpResponse.Code, httpResponse.Body.String())
			}
			savedReceipt, err := newEngine().GetActionReceipt(ctx, campaign, request.ActionID)
			must(err)
			if savedReceipt.Result == nil {
				t.Fatal("receipt reload missing")
			}
			again, err := newEngine().ConfirmAction(ctx, campaign, request)
			must(err)
			a, _ := json.Marshal(first)
			b, _ := json.Marshal(again)
			var va, vb any
			must(json.Unmarshal(a, &va))
			must(json.Unmarshal(b, &vb))
			if fmt.Sprint(va) != fmt.Sprint(vb) {
				t.Fatal("reload replay changed")
			}
			wantCalls := 1
			if choice == "decline" {
				wantCalls = 0
			}
			if dice.calls-failedCalls != wantCalls {
				t.Fatal("replayed dice", dice.calls)
			}
			after, err = q.GetPlayerCharacterByID(ctx, p.ID)
			must(err)
			var saved struct {
				Build  srd521.Build    `json:"srd521"`
				Custom map[string]bool `json:"custom"`
			}
			must(json.Unmarshal(after.Stats, &saved))
			wantUses := 2
			if choice == "success" || choice == "concurrent" {
				wantUses = 1
			}
			if saved.Build.SecondWind != wantUses || !saved.Custom["keep"] || after.Hp != 9 {
				t.Fatalf("%+v", saved)
			}
			if prepare().Available {
				t.Fatal("window reused")
			}
			request.ActionID = uuid.New()
			if _, err := newEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
				t.Fatal(err)
			}
			var original string
			must(testPool.QueryRow(ctx, `SELECT llm_response FROM session_logs WHERE campaign_id=$1 AND turn_number=1`, c.ID).Scan(&original))
			if original != "Original failed-check narration" {
				t.Fatal(original)
			}
		})
	}
}
