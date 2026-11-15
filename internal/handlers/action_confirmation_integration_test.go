//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"sync/atomic"
	"testing"

	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

type confirmationDice struct{ rolls atomic.Int32 }

func (d *confirmationDice) Intn(int) int { d.rolls.Add(1); return 11 }

func TestIntegrationPreparedActionConfirmation(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Confirm", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	campaign, actor := dbutil.FromPgtype(c.ID), dbutil.FromPgtype(pc.ID)
	dice := &confirmationDice{}
	makeEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	prepare := func() engine.PreparedAction {
		campaignData, character, err := game.ReadActionSnapshot(ctx, testPool, campaign, actor, testUserID, false)
		if err != nil {
			t.Fatal(err)
		}
		preview, err := game.PrepareCharacterAction(ctx, campaignData, character, game.ActionIntent{ActorID: actor, Kind: "ability-check", Skill: "athletics"})
		if err != nil || !preview.Available {
			t.Fatalf("preview=%+v %v", preview, err)
		}
		return engine.PreparedAction{ActionID: uuid.New(), Intent: preview.Intent, ExpectedRevision: preview.ExpectedRevision}
	}
	cfg := config.Config{}
	cfg.Server.JWTSecret = "confirmation-test"
	router := server.NewRouterWithProvider(log.New(io.Discard), makeEngine(), q, nil, testPool, testUserID, cfg, nil)
	requestAPI := func(user uuid.UUID, body string) *httptest.ResponseRecorder {
		token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("POST", "/api/v1/campaigns/"+campaign.String()+"/action/confirm", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	if rec := requestAPI(uuid.New(), `{}`); rec.Code != 404 {
		t.Fatalf("owner guard=%d", rec.Code)
	}
	if rec := requestAPI(testUserID, `{"difficulty":10}`); rec.Code != 400 {
		t.Fatalf("client DC accepted=%d", rec.Code)
	}
	request := prepare()
	// Changing HP invalidates the original preview before dice or persistence.
	if _, err := q.UpdatePlayerCurrentHP(ctx, statedb.UpdatePlayerCurrentHPParams{ID: pc.ID, Hp: 8}); err != nil {
		t.Fatal(err)
	}
	if _, err := makeEngine().ConfirmAction(ctx, campaign, request); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("stale error=%v", err)
	}
	request = prepare()
	for _, change := range []string{"actor", "target", "entry", "revision"} {
		invalid := request
		invalid.ActionID = uuid.New()
		switch change {
		case "actor":
			invalid.Intent.ActorID = uuid.New()
		case "target":
			invalid.TargetID = uuid.New()
		case "entry":
			invalid.Intent.EntryID = "attack"
		case "revision":
			invalid.ExpectedRevision = "stale"
		}
		if _, err := makeEngine().ConfirmAction(ctx, campaign, invalid); !errors.Is(err, engine.ErrActionRefresh) {
			t.Fatalf("%s=%v", change, err)
		}
	}
	if dice.rolls.Load() != 0 {
		t.Fatal("invalid confirmation rolled")
	}
	first, err := makeEngine().ConfirmAction(ctx, campaign, request)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		got, err := makeEngine().ConfirmAction(ctx, campaign, request)
		if err != nil || !sameTurnResult(first, got) {
			t.Fatalf("replay=%+v %v", got, err)
		}
	}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	if rec := requestAPI(testUserID, string(body)); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"srd_check"`) {
		t.Fatalf("API replay=%d %s", rec.Code, rec.Body.String())
	}
	// Reconnect over a real WebSocket and retry the same typed confirmation.
	srv := httptest.NewServer(router)
	defer srv.Close()
	token, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	socketCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	conn, _, err := websocket.Dial(socketCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/campaigns/"+campaign.String()+"/ws", nil)
	if err == nil {
		_ = conn.CloseNow()
		t.Fatal("unauthorized WebSocket admitted")
	}
	headers := make(http.Header)
	headers.Set("Authorization", "Bearer "+token)
	conn, _, err = websocket.Dial(socketCtx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/api/v1/campaigns/"+campaign.String()+"/ws", &websocket.DialOptions{HTTPHeader: headers})
	if err != nil {
		t.Fatal(err)
	}
	defer conn.CloseNow()
	if err := wsjson.Write(socketCtx, conn, map[string]any{"type": "confirm_action", "payload": request}); err != nil {
		t.Fatal(err)
	}
	var envelope api.WebSocketMessageEnvelope
	if err := wsjson.Read(socketCtx, conn, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Type != "result" || !strings.Contains(string(envelope.Payload), request.ActionID.String()) {
		t.Fatalf("WS replay=%+v", envelope)
	}
	conflict := request
	conflict.Intent.Skill = "perception"
	conflict.Intent.Ability = "wisdom"
	if _, err := makeEngine().ConfirmAction(ctx, campaign, conflict); !errors.Is(err, engine.ErrActionIDConflict) {
		t.Fatalf("conflict=%v", err)
	}
	if _, err := makeEngine().ConfirmAction(ctx, uuid.New(), request); !errors.Is(err, engine.ErrActionIDConflict) {
		t.Fatalf("campaign conflict=%v", err)
	}
	stale := request
	stale.ActionID = uuid.New()
	if _, err := makeEngine().ConfirmAction(ctx, campaign, stale); !errors.Is(err, engine.ErrActionRefresh) {
		t.Fatalf("history did not invalidate preview: %v", err)
	}
	receipt, err := makeEngine().GetActionReceipt(ctx, campaign, request.ActionID)
	if err != nil || receipt.Status != "completed" || !sameTurnResult(first, receipt.Result) {
		t.Fatalf("receipt=%+v %v", receipt, err)
	}
	var result struct {
		Receipt struct {
			Total   int    `json:"total"`
			Outcome string `json:"outcome"`
		} `json:"receipt"`
	}
	if err := json.Unmarshal(first.AppliedToolCalls[0].Result, &result); err != nil {
		t.Fatal(err)
	}
	if result.Receipt.Total != 17 || result.Receipt.Outcome != "unknown" || dice.rolls.Load() != 1 {
		t.Fatalf("result=%+v rolls=%d", result, dice.rolls.Load())
	}
	logs, err := q.ListSessionLogsByCampaign(ctx, c.ID)
	if err != nil || len(logs) != 1 {
		t.Fatalf("history=%d %v", len(logs), err)
	}
	stored, err := q.GetPlayerCharacterByID(ctx, pc.ID)
	if err != nil || stored.Hp != 8 {
		t.Fatalf("resource changed %+v %v", stored, err)
	}
}
