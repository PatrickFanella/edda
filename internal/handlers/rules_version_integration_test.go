//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/llm"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/tools"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

func TestIntegrationRulesVersionCompatibility(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	cfg := config.Config{}
	cfg.Server.JWTSecret = "rules-version-integration"
	router := server.NewRouterWithProvider(log.New(io.Discard), &stubIntegrationEngine{}, q, nil, testPool, testUserID, cfg, saves.NewStore(testPool))
	token, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	request := func(method, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest(method, "/api/v1/campaigns"+path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, r)
		if rec.Code != want {
			t.Fatalf("%s %s: %d %s, want %d", method, path, rec.Code, rec.Body.String(), want)
		}
		return rec
	}
	for _, mode := range []string{"narrative", "light", "crunch"} {
		t.Run(mode, func(t *testing.T) {
			rec := request("POST", "", `{"name":"Legacy","rules_mode":"`+mode+`"}`, http.StatusCreated)
			var c api.CampaignResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &c); err != nil {
				t.Fatal(err)
			}
			if c.RulesVersion != "legacy" || c.RulesMode != mode {
				t.Fatalf("legacy metadata = %+v", c)
			}
			// Existing mechanics and stored character data remain usable.
			id := uuid.MustParse(c.ID)
			pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: dbutil.ToPgtype(id), UserID: dbutil.ToPgtype(testUserID), Name: "Legacy hero", Stats: []byte(`{"strength":12}`), Hp: 9, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
			if err != nil {
				t.Fatal(err)
			}
			request("PUT", "/"+c.ID, `{"name":"Legacy renamed"}`, http.StatusOK)
			got, err := q.GetPlayerCharacterByID(ctx, pc.ID)
			if err != nil || got.Hp != 9 || string(got.Stats) != `{"strength": 12}` {
				t.Fatalf("legacy character changed: %+v %v", got, err)
			}
			request("PUT", "/"+c.ID, `{"name":"Convert","rules_version":"srd-5.2.1"}`, http.StatusConflict)
		})
	}
	request("POST", "", `{"name":"Unknown","rules_version":"srd-unknown"}`, http.StatusBadRequest)
	rec := request("POST", "", `{"name":"SRD metadata","rules_mode":"light","rules_version":"srd-5.2.1"}`, http.StatusCreated)
	var c api.CampaignResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	if c.RulesVersion != "srd-5.2.1" || c.RulesMode != "light" {
		t.Fatalf("SRD metadata = %+v", c)
	}
	request("GET", "/"+c.ID, "", http.StatusOK)
	request("GET", "", "", http.StatusOK)
	for _, mutation := range []struct{ method, path, body string }{
		{"POST", "/action", `{"input":"level up"}`},
		{"PUT", "", `{"name":"Changed","rules_version":"legacy"}`},
		{"POST", "/start-over", ""},
		{"POST", "/saves", `{"name":"old client save"}`},
		{"POST", "/journal/entries", `{"title":"old client","content":"write"}`},
		{"DELETE", "", ""},
	} {
		request(mutation.method, "/"+c.ID+mutation.path, mutation.body, http.StatusConflict)
	}
	// A legacy WebSocket handshake cannot enter the action transport either.
	wsRequest := httptest.NewRequest("GET", "/api/v1/campaigns/"+c.ID+"/ws", nil)
	wsRequest.Header.Set("Authorization", "Bearer "+token)
	wsRequest.Header.Set("Upgrade", "websocket")
	wsRequest.Header.Set("Connection", "Upgrade")
	wsResponse := httptest.NewRecorder()
	router.ServeHTTP(wsResponse, wsRequest)
	if wsResponse.Code != http.StatusConflict {
		t.Fatalf("websocket = %d %s", wsResponse.Code, wsResponse.Body.String())
	}
	id := uuid.MustParse(c.ID)
	state, err := game.NewStateManager(testPool).GatherState(ctx, id)
	if err != nil || state.Campaign.RulesVersion != domain.RulesVersionSRD521 {
		t.Fatalf("domain metadata = %+v, %v", state, err)
	}
	// The same engine entry points used by the local TUI also refuse writes.
	realEngine, err := engine.New(testPool, nil, config.LLMConfig{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := realEngine.ProcessTurn(ctx, id, "level up"); !errors.Is(err, domain.ErrLegacyRulesMutation) {
		t.Fatalf("TUI turn = %v", err)
	}
	stream, err := realEngine.ProcessTurnStream(ctx, id, "level up")
	if err != nil {
		t.Fatal(err)
	}
	var streamErr error
	for ev := range stream {
		if ev.Type == "error" {
			streamErr = ev.Err
		}
	}
	if !errors.Is(streamErr, domain.ErrLegacyRulesMutation) {
		t.Fatalf("stream turn = %v", streamErr)
	}
	// Tool execution checks persisted metadata even without the engine pipeline.
	registry := tools.NewRegistry()
	registry.RequireLegacyCampaign(q)
	called := false
	if err := registry.Register(llm.Tool{Name: "legacy_mutation"}, func(context.Context, map[string]any) (*tools.ToolResult, error) {
		called = true
		return &tools.ToolResult{Success: true}, nil
	}); err != nil {
		t.Fatal(err)
	}
	_, err = registry.Execute(tools.WithCurrentCampaignID(ctx, id), "legacy_mutation", nil)
	if !errors.Is(err, domain.ErrLegacyRulesMutation) || called {
		t.Fatalf("tool mutation = %v, called=%v", err, called)
	}
	if _, err := registry.Execute(ctx, "legacy_mutation", nil); err == nil || called {
		t.Fatal("tool allowed missing campaign")
	}
	legacy, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Legacy tool", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Execute(tools.WithCurrentCampaignID(ctx, dbutil.FromPgtype(legacy.ID)), "legacy_mutation", nil); err != nil || !called {
		t.Fatalf("legacy tool = %v called=%v", err, called)
	}
	logs, err := q.ListSessionLogsByCampaign(ctx, dbutil.ToPgtype(id))
	if err != nil || len(logs) != 0 {
		t.Fatalf("refused turns saved history: %+v %v", logs, err)
	}
}
