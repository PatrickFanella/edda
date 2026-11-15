//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

type narrationRecoveryProvider struct {
	t     *testing.T
	calls int
	fail  bool
	tools bool
}

func (p *narrationRecoveryProvider) Complete(ctx context.Context, messages []llm.Message, available []llm.Tool) (*llm.Response, error) {
	p.calls++
	if len(available) != 0 {
		p.t.Fatal("recovery advertised mechanical tools")
	}
	var open int
	if err := testPool.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction'`).Scan(&open); err != nil {
		p.t.Fatal(err)
	}
	if open != 0 {
		p.t.Fatalf("inference holds %d transactions", open)
	}
	prompt := messages[len(messages)-1].Content
	if !strings.Contains(prompt, `"roll":17`) || !strings.Contains(prompt, `"hp":7`) {
		p.t.Fatalf("missing exact receipts: %s", prompt)
	}
	if p.fail {
		return nil, errors.New("injected provider failure")
	}
	response := &llm.Response{Content: "The recorded result holds.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen."}
	if p.tools {
		response.ToolCalls = []llm.ToolCall{{Name: "update_player_hp", Arguments: map[string]any{"hp": 1}}}
	}
	return response, nil
}
func (p *narrationRecoveryProvider) Stream(ctx context.Context, messages []llm.Message, available []llm.Tool) (<-chan llm.StreamChunk, error) {
	return streamReceiptResponse(p.Complete(ctx, messages, available))
}

func seedNarrationRecovery(t *testing.T, status string) (uuid.UUID, uuid.UUID, statedb.PlayerCharacter) {
	t.Helper()
	ctx := context.Background()
	q := statedb.New(testPool)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Narration recovery", Status: "active", CreatedBy: db.ToPgUUID(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: db.ToPgUUID(testUserID), Name: "Receipt hero", Stats: []byte(`{}`), Hp: 7, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	id := uuid.New()
	calls := fmt.Sprintf(`[{"Tool":"skill_check","Arguments":{"character_id":"%s","skill":"athletics","difficulty":12},"Result":{"roll":17,"rolls":[17],"modifier":2,"total":19,"success":true}},{"Tool":"update_player_hp","Arguments":{"hp":7},"Result":{"hp":7}}]`, db.FromPgUUID(pc.ID))
	_, err = testPool.Exec(ctx, `INSERT INTO turn_attempts(id,campaign_id,turn_number,player_input,status,calls,candidate_narrative) VALUES($1,$2,1,'Climb',$3,$4,'Unvalidated raw candidate')`, db.ToPgUUID(id), c.ID, status, []byte(calls))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(context.Background(), `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	return db.FromPgUUID(c.ID), id, pc
}

func TestIntegrationNarrationRecoveryAfterRestart(t *testing.T) {
	for _, status := range []string{"pending_mechanics", "pending_narration"} {
		t.Run(status, func(t *testing.T) {
			ctx := context.Background()
			campaign, action, pc := seedNarrationRecovery(t, status)
			provider := &narrationRecoveryProvider{t: t, fail: true}
			makeEngine := func() *engine.Engine {
				e, err := engine.New(testPool, provider, config.LLMConfig{})
				if err != nil {
					t.Fatal(err)
				}
				return e
			}
			e := makeEngine()
			// Inspect and discover through the production owner-guarded API.
			cfg := config.Config{}
			cfg.Server.JWTSecret = "narration-recovery-fixture"
			router := server.NewRouterWithProvider(log.New(io.Discard), e, statedb.New(testPool), provider, testPool, testUserID, cfg, nil)
			read := func(user uuid.UUID, path string) *httptest.ResponseRecorder {
				token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
				if err != nil {
					t.Fatal(err)
				}
				req := httptest.NewRequest("GET", path, nil)
				req.Header.Set("Authorization", "Bearer "+token)
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, req)
				return rec
			}
			path := "/api/v1/campaigns/" + campaign.String() + "/action"
			pending := read(testUserID, path)
			var pendingReceipt api.ActionReceiptResponse
			if err := json.Unmarshal(pending.Body.Bytes(), &pendingReceipt); err != nil || pending.Code != 200 || pendingReceipt.Status != "pending_narration" || pendingReceipt.Input != "Climb" || len(pendingReceipt.ResolutionEvents) != 1 || pendingReceipt.Result != nil || strings.Contains(pending.Body.String(), "Unvalidated raw") {
				t.Fatalf("pending API = %d %s error=%v", pending.Code, pending.Body.String(), err)
			}

			before, err := e.GetActionReceipt(ctx, campaign, uuid.Nil)
			if err != nil || before == nil || before.Status != "pending_narration" || len(before.AppliedToolCalls) != 2 {
				t.Fatalf("pending receipt: %+v %v", before, err)
			}
			if _, err := e.ProcessTurn(engine.WithActionID(ctx, action), campaign, "Climb"); !errors.Is(err, engine.ErrTurnRecoveryRequired) {
				t.Fatalf("failure=%v", err)
			}
			after, err := e.GetActionReceipt(ctx, campaign, action)
			if err != nil || after.Status != "pending_narration" || after.NarrationAttempts != 1 {
				t.Fatalf("failed retry receipt=%+v %v", after, err)
			}
			provider.fail = false
			e = makeEngine() // A new process has no in-memory journal or dice state.
			events, err := e.ProcessTurnStream(engine.WithActionID(ctx, action), campaign, "Climb")
			if err != nil {
				t.Fatal(err)
			}
			var result *engine.TurnResult
			var preview string
			for event := range events {
				if event.Type == "chunk" {
					preview += event.Text
				}
				if event.Err != nil {
					t.Fatal(event.Err)
				}
				if event.Result != nil {
					result = event.Result
				}
			}
			if !strings.Contains(preview, "The recorded result holds.") {
				t.Fatal("recovery did not stream provisional narration")
			}
			if result == nil || !result.Recovered || len(result.AppliedToolCalls) != 2 || len(result.Choices) != 3 || strings.Contains(result.Narrative, "Mechanical outcomes") {
				t.Fatalf("recovery=%+v", result)
			}
			if !sameTurnResult(&engine.TurnResult{AppliedToolCalls: before.AppliedToolCalls}, &engine.TurnResult{AppliedToolCalls: result.AppliedToolCalls}) {
				t.Fatal("recovery changed the recorded dice/effect receipts")
			}
			calls := provider.calls
			again, err := makeEngine().ProcessTurn(engine.WithActionID(ctx, action), campaign, "Climb")
			if err != nil || !sameTurnResult(result, again) || provider.calls != calls {
				t.Fatalf("replay changed result: %+v %v", again, err)
			}
			got, err := statedb.New(testPool).GetPlayerCharacterByID(ctx, pc.ID)
			if err != nil || got.Hp != 7 {
				t.Fatalf("effects repeated: %+v %v", got, err)
			}
			logs, err := statedb.New(testPool).ListSessionLogsByCampaign(ctx, db.ToPgUUID(campaign))
			if err != nil || len(logs) != 1 {
				t.Fatalf("history=%+v %v", logs, err)
			}

			rec := read(testUserID, path+"?action_id="+action.String())
			var receipt api.ActionReceiptResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &receipt); err != nil || rec.Code != 200 || receipt.Status != "completed" || receipt.Result == nil || receipt.Result.Narrative != engine.CleanNarrativeText(result.Narrative) {
				t.Fatalf("API=%d %s error=%v", rec.Code, rec.Body.String(), err)
			}
			if rec := read(testUserID, path); rec.Code != 200 || strings.TrimSpace(rec.Body.String()) != "null" {
				t.Fatalf("pending discovery=%d %s", rec.Code, rec.Body.String())
			}
			if rec := read(uuid.New(), path); rec.Code != 404 {
				t.Fatalf("receipt leaked across owner: %d", rec.Code)
			}
			if rec := read(testUserID, path+"?action_id="+uuid.NewString()); strings.TrimSpace(rec.Body.String()) != "null" {
				t.Fatalf("wrong id leaked receipt: %s", rec.Body.String())
			}
		})
	}
}

func TestIntegrationNarrationRecoveryBoundedAcrossRestarts(t *testing.T) {
	for _, invalidTool := range []bool{false, true} {
		t.Run(fmt.Sprint(invalidTool), func(t *testing.T) {
			campaign, action, _ := seedNarrationRecovery(t, "pending_narration")
			provider := &narrationRecoveryProvider{t: t, fail: !invalidTool, tools: invalidTool}
			for attempt := 0; attempt < engine.MaxNarrationAttempts+2; attempt++ {
				e, err := engine.New(testPool, provider, config.LLMConfig{})
				if err != nil {
					t.Fatal(err)
				}
				_, err = e.ProcessTurn(engine.WithActionID(context.Background(), action), campaign, "Climb")
				if !errors.Is(err, engine.ErrTurnRecoveryRequired) {
					t.Fatalf("retry %d error=%v", attempt, err)
				}
				if attempt >= engine.MaxNarrationAttempts && !strings.Contains(err.Error(), "attempt limit") {
					t.Fatalf("limit not explicit: %v", err)
				}
			}
			if provider.calls != engine.MaxNarrationAttempts {
				t.Fatalf("unbounded inference calls=%d", provider.calls)
			}
			var status string
			var count int
			if err := testPool.QueryRow(context.Background(), `SELECT status,narration_attempts FROM turn_attempts WHERE id=$1`, db.ToPgUUID(action)).Scan(&status, &count); err != nil {
				t.Fatal(err)
			}
			if status != "pending_narration" || count != engine.MaxNarrationAttempts {
				t.Fatalf("lost pending receipt: %s %d", status, count)
			}
			var logs int
			if err := testPool.QueryRow(context.Background(), `SELECT count(*) FROM session_logs WHERE campaign_id=$1`, db.ToPgUUID(campaign)).Scan(&logs); err != nil || logs != 0 {
				t.Fatalf("failed model leaked into history: %d %v", logs, err)
			}
		})
	}
}
