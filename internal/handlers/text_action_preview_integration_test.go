//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
)

func TestIntegrationAthleticsTextPreview(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	campaign, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Text preview", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(context.Background(), `DELETE FROM campaigns WHERE id=$1`, campaign.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: campaign.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 12, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	campaignID, actorID := dbutil.FromPgtype(campaign.ID), dbutil.FromPgtype(pc.ID)
	dice := &confirmationDice{}
	newEngine := func() *engine.Engine {
		e, err := engine.New(testPool, nil, config.LLMConfig{}, engine.WithActionDice(dice))
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	cfg := config.Config{}
	cfg.Server.JWTSecret = "text-preview-test"
	router := server.NewRouterWithProvider(log.New(io.Discard), newEngine(), q, nil, testPool, testUserID, cfg, nil)
	path := "/api/v1/campaigns/" + campaignID.String() + "/character/action-preview"
	request := func(user uuid.UUID, method, body, contentType string) *httptest.ResponseRecorder {
		t.Helper()
		token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	snapshot := func() string {
		t.Helper()
		var raw string
		if err := testPool.QueryRow(ctx, `SELECT jsonb_build_object('campaign',to_jsonb(c),'characters',(SELECT jsonb_agg(p) FROM player_characters p WHERE p.campaign_id=c.id),'clock',(SELECT jsonb_agg(t) FROM campaign_time t WHERE t.campaign_id=c.id),'logs',(SELECT jsonb_agg(l) FROM session_logs l WHERE l.campaign_id=c.id),'attempts',(SELECT jsonb_agg(a) FROM turn_attempts a WHERE a.campaign_id=c.id))::text FROM campaigns c WHERE c.id=$1`, campaign.ID).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		return raw
	}

	before := snapshot()
	payload := `{"actor_id":"` + actorID.String() + `","input":"Make an Athletics check"}`
	for _, body := range []string{
		`{}`, `{"actor_id":"` + actorID.String() + `","input":"Inspect the gate"}`,
		`{"actor_id":"` + actorID.String() + `","input":" Make an Athletics check"}`,
		`{"actor_id":"` + actorID.String() + `","input":"Make an athletics check"}`,
		`{"actor_id":"` + actorID.String() + `","input":"Make an Athletics check","difficulty":10}`,
		`{"actor_id":"` + actorID.String() + `","input":"Make an Athletics check","kind":"ability-check","skill":"athletics"}`,
		payload + ` {}`, `{"actor_id":"` + actorID.String() + `","input":"` + strings.Repeat("x", 2048) + `"}`,
	} {
		rec := request(testUserID, http.MethodPost, body, "application/json")
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Use the exact supported text") || rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatalf("invalid input status=%d headers=%v body=%s", rec.Code, rec.Header(), rec.Body.String())
		}
		if snapshot() != before {
			t.Fatal("invalid text preview changed persistent state")
		}
	}
	if rec := request(testUserID, http.MethodPost, payload, "text/plain"); rec.Code != http.StatusBadRequest || snapshot() != before {
		t.Fatalf("content type status=%d body=%s", rec.Code, rec.Body.String())
	}
	token, err := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	queryReq := httptest.NewRequest(http.MethodPost, path+"?skill=athletics", strings.NewReader(payload))
	queryReq.Header.Set("Authorization", "Bearer "+token)
	queryReq.Header.Set("Content-Type", "application/json")
	queryRec := httptest.NewRecorder()
	router.ServeHTTP(queryRec, queryReq)
	if queryRec.Code != http.StatusBadRequest || !strings.Contains(queryRec.Body.String(), "Use the exact supported text") || queryRec.Header().Get("Cache-Control") != "no-store" || snapshot() != before {
		t.Fatalf("query status=%d headers=%v body=%s", queryRec.Code, queryRec.Header(), queryRec.Body.String())
	}
	if rec := request(uuid.New(), http.MethodPost, payload, "application/json"); rec.Code != http.StatusNotFound || snapshot() != before {
		t.Fatalf("foreign actor status=%d body=%s", rec.Code, rec.Body.String())
	}
	for _, mutation := range []struct{ path, body string }{
		{"/action", `{"input":"Inspect the gate"}`},
		{"/saves", `{"name":"blocked"}`},
		{"/start-over", ""},
	} {
		token, tokenErr := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
		if tokenErr != nil {
			t.Fatal(tokenErr)
		}
		req := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+campaignID.String()+mutation.path, strings.NewReader(mutation.body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusConflict || snapshot() != before {
			t.Fatalf("legacy guard %s status=%d body=%s", mutation.path, rec.Code, rec.Body.String())
		}
	}

	textRec := request(testUserID, http.MethodPost, payload, "application/json")
	if textRec.Code != http.StatusOK || textRec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("text preview status=%d headers=%v body=%s", textRec.Code, textRec.Header(), textRec.Body.String())
	}
	var textPreview api.ActionPreview
	if err := json.Unmarshal(textRec.Body.Bytes(), &textPreview); err != nil {
		t.Fatal(err)
	}
	if !textPreview.Available || textPreview.Intent.Kind != "ability-check" || textPreview.Intent.Skill != "athletics" || textPreview.Intent.Ability != "strength" {
		t.Fatalf("text preview=%+v", textPreview)
	}
	typedRec := func() *httptest.ResponseRecorder {
		token, tokenErr := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
		if tokenErr != nil {
			t.Fatal(tokenErr)
		}
		req := httptest.NewRequest(http.MethodGet, path+"?actor_id="+actorID.String()+"&kind=ability-check&skill=athletics", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}()
	var typedPreview api.ActionPreview
	if typedRec.Code != http.StatusOK || json.Unmarshal(typedRec.Body.Bytes(), &typedPreview) != nil || !reflect.DeepEqual(textPreview, typedPreview) || snapshot() != before {
		t.Fatalf("typed status=%d text=%+v typed=%+v", typedRec.Code, textPreview, typedPreview)
	}

	prepared := engine.PreparedAction{ActionID: uuid.New(), ExpectedRevision: textPreview.ExpectedRevision, Intent: textPreview.Intent}
	body, err := json.Marshal(prepared)
	if err != nil {
		t.Fatal(err)
	}
	confirmPath := "/api/v1/campaigns/" + campaignID.String() + "/action/confirm"
	confirm := func() *httptest.ResponseRecorder {
		token, tokenErr := auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
		if tokenErr != nil {
			t.Fatal(tokenErr)
		}
		req := httptest.NewRequest(http.MethodPost, confirmPath, strings.NewReader(string(body)))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	first, replay := confirm(), confirm()
	if first.Code != http.StatusOK || replay.Code != http.StatusOK || first.Body.String() != replay.Body.String() || dice.rolls.Load() != 1 {
		t.Fatalf("confirm=%d replay=%d rolls=%d", first.Code, replay.Code, dice.rolls.Load())
	}
	receipt, err := newEngine().GetActionReceipt(ctx, campaignID, prepared.ActionID)
	if err != nil || receipt == nil || receipt.Status != "completed" {
		t.Fatalf("receipt=%+v err=%v", receipt, err)
	}
	logs, err := q.ListSessionLogsByCampaign(ctx, campaign.ID)
	if err != nil || len(logs) != 1 {
		t.Fatalf("history=%d err=%v", len(logs), err)
	}

	unsupported, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Unsupported preview", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = testPool.Exec(context.Background(), `DELETE FROM campaigns WHERE id=$1`, unsupported.ID)
	})
	unsupportedPC, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: unsupported.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Unsupported", Stats: []byte(`{}`), Hp: 10, MaxHp: 10, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	token, err = auth.GenerateToken(testUserID, cfg.Server.JWTSecret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	unsupportedReq := httptest.NewRequest(http.MethodPost, "/api/v1/campaigns/"+dbutil.FromPgtype(unsupported.ID).String()+"/character/action-preview", strings.NewReader(`{"actor_id":"`+dbutil.FromPgtype(unsupportedPC.ID).String()+`","input":"Make an Athletics check"}`))
	unsupportedReq.Header.Set("Authorization", "Bearer "+token)
	unsupportedReq.Header.Set("Content-Type", "application/json")
	unsupportedRec := httptest.NewRecorder()
	router.ServeHTTP(unsupportedRec, unsupportedReq)
	var unavailable api.ActionPreview
	if unsupportedRec.Code != http.StatusOK || json.Unmarshal(unsupportedRec.Body.Bytes(), &unavailable) != nil || unavailable.Available || unavailable.Reason == "" {
		t.Fatalf("unsupported status=%d preview=%+v body=%s", unsupportedRec.Code, unavailable, unsupportedRec.Body.String())
	}
}
