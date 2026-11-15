//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"io"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

type receiptProvider struct {
	t         *testing.T
	calls     int
	noTools   bool
	noChoices bool
	multi     bool
}

func (p *receiptProvider) Complete(ctx context.Context, messages []llm.Message, available []llm.Tool) (*llm.Response, error) {
	var open int
	if err := testPool.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND state='idle in transaction'`).Scan(&open); err != nil {
		p.t.Fatal(err)
	}
	if open != 0 {
		p.t.Errorf("provider called with %d open transactions", open)
	}
	p.calls++
	if len(available) == 0 && strings.Contains(messages[len(messages)-1].Content, "Generate 3-5 player choices") {
		return &llm.Response{Content: "1. Wait.\n2. Look.\n3. Listen."}, nil
	}
	if len(available) == 0 {
		return &llm.Response{Content: "A quiet pause.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen."}, nil
	}
	if p.noChoices {
		return &llm.Response{Content: "A quiet pause.", ToolCalls: []llm.ToolCall{{ID: "hp", Name: "update_player_hp", Arguments: map[string]any{"hp": 7}}}}, nil
	}
	if p.noTools {
		return &llm.Response{Content: "A quiet pause.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen."}, nil
	}
	if p.multi {
		return &llm.Response{Content: "A quiet pause.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen.", ToolCalls: []llm.ToolCall{{ID: "first", Name: "update_player_hp", Arguments: map[string]any{"hp": 7}}, {ID: "second", Name: "update_player_hp", Arguments: map[string]any{"hp": 6}}}}, nil
	}
	return &llm.Response{Content: "A quiet pause.\n\nChoices:\n1. Wait.\n2. Look.\n3. Listen.", ToolCalls: []llm.ToolCall{{ID: "hp", Name: "update_player_hp", Arguments: map[string]any{"hp": 7}}}}, nil
}
func (*receiptProvider) Stream(context.Context, []llm.Message, []llm.Tool) (<-chan llm.StreamChunk, error) {
	return nil, errors.New("unused")
}
func TestIntegrationDurableTurnCompletion(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	create := func(name string) (uuid.UUID, statedb.PlayerCharacter) {
		c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: name, Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
		if err != nil {
			t.Fatal(err)
		}
		pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Receipt hero", Stats: []byte(`{}`), Hp: 10, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _, _ = testPool.Exec(context.Background(), `DELETE FROM campaigns WHERE id=$1`, c.ID) })
		return dbutil.FromPgtype(c.ID), pc
	}
	for _, point := range []string{"completion", "completion-no-choices", "receipt", "zero-tools", "second-receipt"} {
		t.Run(point, func(t *testing.T) {
			id, pc := create(point)
			actionID := uuid.New()
			p := &receiptProvider{t: t, noTools: point == "zero-tools", noChoices: point == "completion-no-choices", multi: point == "second-receipt"}
			e, err := engine.New(testPool, p, config.LLMConfig{})
			if err != nil {
				t.Fatal(err)
			}
			table, condition := "session_logs", "NEW.llm_response NOT LIKE 'Mechanical outcomes%'"
			if point == "receipt" || point == "second-receipt" {
				table = "turn_attempts"
				condition = "NEW.calls <> '[]'::jsonb"
			}
			if point == "second-receipt" {
				condition = "jsonb_array_length(NEW.calls) = 2"
			}
			function := "fail_turn_" + strings.ReplaceAll(point, "-", "_")
			sql := fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.campaign_id='%s' AND %s THEN RAISE EXCEPTION 'injected turn persistence failure'; END IF; RETURN NEW; END; $$; CREATE TRIGGER %s BEFORE INSERT OR UPDATE ON %s FOR EACH ROW EXECUTE FUNCTION %s()`, function, id, condition, function, table, function)
			if _, err := testPool.Exec(ctx, sql); err != nil {
				t.Fatal(err)
			}
			cleanup := func() {
				_, err := testPool.Exec(ctx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON %s; DROP FUNCTION IF EXISTS %s()`, function, table, function))
				if err != nil {
					t.Error(err)
				}
			}
			defer cleanup()
			if point == "completion" || point == "completion-no-choices" {
				cfg := config.Config{}
				cfg.Server.JWTSecret = "turn-completion-fixture"
				router := server.NewRouterWithProvider(log.New(io.Discard), e, q, p, testPool, testUserID, cfg, saves.NewStore(testPool))
				token, tokenErr := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
				if tokenErr != nil {
					t.Fatal(tokenErr)
				}
				request := httptest.NewRequest("POST", "/api/v1/campaigns/"+id.String()+"/action", strings.NewReader(fmt.Sprintf(`{"input":"Wait here","action_id":"%s"}`, actionID)))
				request.Header.Set("Authorization", "Bearer "+token)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				var payload map[string]string
				_ = json.Unmarshal(response.Body.Bytes(), &payload)
				if response.Code != 409 || payload["action_id"] != actionID.String() {
					t.Fatalf("recovery API = %d %s", response.Code, response.Body.String())
				}
				err = engine.ErrTurnRecoveryRequired
			} else if point == "second-receipt" {
				var events <-chan engine.StreamEvent
				events, err = e.ProcessTurnStream(engine.WithActionID(ctx, actionID), id, "Wait here")
				if err != nil {
					t.Fatal(err)
				}
				for event := range events {
					if event.Err != nil {
						err = event.Err
					}
				}
			} else {
				_, err = e.ProcessTurn(engine.WithActionID(ctx, actionID), id, "Wait here")
			}
			if !errors.Is(err, engine.ErrTurnRecoveryRequired) {
				t.Fatalf("failure = %v", err)
			}
			logs, err := q.ListSessionLogsByCampaign(ctx, dbutil.ToPgtype(id))
			if err != nil || len(logs) != 0 {
				t.Fatalf("pending attempt leaked into history: %+v %v", logs, err)
			}
			got, err := q.GetPlayerCharacterByID(ctx, pc.ID)
			if err != nil {
				t.Fatal(err)
			}
			want := int32(7)
			if point == "receipt" || point == "zero-tools" {
				want = 10
			}
			if got.Hp != want {
				t.Fatalf("HP %d want %d", got.Hp, want)
			}
			callsBefore := p.calls
			if point == "completion" || point == "zero-tools" || point == "second-receipt" {
				if _, err := e.ProcessTurn(engine.WithActionID(ctx, uuid.New()), id, "Different action"); !errors.Is(err, engine.ErrTurnRecoveryRequired) {
					t.Fatalf("new action bypassed recovery: %v", err)
				}
			}
			cleanup()
			result, err := e.ProcessTurn(engine.WithActionID(ctx, actionID), id, "Wait here")
			if err != nil {
				t.Fatal(err)
			}
			if (point == "completion" || point == "completion-no-choices" || point == "zero-tools") && (!result.Recovered || p.calls != callsBefore) {
				t.Fatalf("retry reran committed work: %+v calls %d -> %d", result, callsBefore, p.calls)
			}
			if point == "completion-no-choices" && len(result.Choices) != 3 {
				t.Fatalf("recovery lost generated choices: %+v", result.Choices)
			}
			callsAfter := p.calls
			again, err := e.ProcessTurn(engine.WithActionID(ctx, actionID), id, "Wait here")
			if err != nil || !sameTurnResult(again, result) || p.calls != callsAfter {
				t.Fatalf("completed retry = %+v %v calls %d", again, err, p.calls)
			}
			if _, err := e.ProcessTurn(engine.WithActionID(ctx, actionID), id, "Different input"); err == nil {
				t.Fatal("same id accepted different input")
			}
			logs, err = q.ListSessionLogsByCampaign(ctx, dbutil.ToPgtype(id))
			if err != nil || len(logs) != 1 {
				t.Fatalf("completed history = %+v %v", logs, err)
			}
			var candidate, status string
			if err := testPool.QueryRow(ctx, `SELECT candidate_narrative,status FROM turn_attempts WHERE id=$1`, dbutil.ToPgtype(actionID)).Scan(&candidate, &status); err != nil {
				t.Fatal(err)
			}
			if candidate == "" || status != "completed" {
				t.Fatalf("lost candidate/lifecycle: %q %q", candidate, status)
			}
			// The same text with a fresh ID is a deliberate new action.
			if _, err := e.ProcessTurn(engine.WithActionID(ctx, uuid.New()), id, "Wait here"); err != nil {
				t.Fatal(err)
			}
			if p.calls == callsAfter {
				t.Fatal("new repeated input was incorrectly deduplicated")
			}
			replay, err := e.ProcessTurn(engine.WithActionID(ctx, actionID), id, "Wait here")
			if err != nil || !sameTurnResult(replay, result) {
				t.Fatalf("historical replay changed outcome: %+v %v", replay, err)
			}
			other, _ := create("Other owner boundary")
			if _, err := e.ProcessTurn(engine.WithActionID(ctx, actionID), other, "Wait here"); err == nil {
				t.Fatal("cross-campaign action id reused")
			}
		})
	}
}

// PostgreSQL JSONB normalizes RawMessage whitespace/key order. Compare the
// decoded response values, retaining every field and mechanical receipt.
func sameTurnResult(a, b *engine.TurnResult) bool {
	left, leftErr := json.Marshal(a)
	right, rightErr := json.Marshal(b)
	if leftErr != nil || rightErr != nil {
		return false
	}
	var leftValue, rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	return reflect.DeepEqual(leftValue, rightValue)
}
