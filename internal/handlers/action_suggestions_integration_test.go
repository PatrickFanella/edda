//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationGroundedSuggestions(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Suggestions", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	for _, fact := range []struct {
		text  string
		known bool
	}{{"Blue seal on gate", true}, {"Secret door behind throne", false}} {
		_, err := q.CreateFact(ctx, statedb.CreateFactParams{CampaignID: c.ID, Fact: fact.text, Category: "clue", Source: "fixture", PlayerKnown: fact.known})
		if err != nil {
			t.Fatal(err)
		}
	}
	cfg := config.Config{}
	cfg.Server.JWTSecret = "suggestion-test"
	router := server.NewRouterWithProvider(log.New(io.Discard), nil, q, nil, testPool, testUserID, cfg, nil)
	request := func(user uuid.UUID) *httptest.ResponseRecorder {
		token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/api/v1/campaigns/"+dbutil.FromPgtype(c.ID).String()+"/action-suggestions", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	if rec := request(uuid.New()); rec.Code != 404 {
		t.Fatalf("foreign owner=%d", rec.Code)
	}
	loc, err := q.CreateLocation(ctx, statedb.CreateLocationParams{CampaignID: c.ID, Name: "Hidden Sanctum", Description: pgtype.Text{String: "Private room", Valid: true}, LocationType: pgtype.Text{String: "building", Valid: true}, Properties: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.UpdatePlayerLocation(ctx, statedb.UpdatePlayerLocationParams{ID: pc.ID, CurrentLocationID: loc.ID}); err != nil {
		t.Fatal(err)
	}
	first := request(testUserID)
	if first.Code != 200 || strings.Contains(first.Body.String(), "Hidden Sanctum") || strings.Contains(first.Body.String(), "Secret") || !strings.Contains(first.Body.String(), "Blue seal") || strings.Contains(first.Body.String(), "second-wind") {
		t.Fatalf("unsafe suggestions=%d %s", first.Code, first.Body.String())
	}
	var before game.ActionSuggestions
	if err := json.Unmarshal(first.Body.Bytes(), &before); err != nil {
		t.Fatal(err)
	}
	if err := q.SetLocationPlayerKnown(ctx, loc.ID); err != nil {
		t.Fatal(err)
	}
	sceneResponse := request(testUserID)
	var scene game.ActionSuggestions
	if err := json.Unmarshal(sceneResponse.Body.Bytes(), &scene); err != nil {
		t.Fatal(err)
	}
	if scene.Version == before.Version || !strings.Contains(sceneResponse.Body.String(), "Look around Hidden Sanctum") {
		t.Fatal("known scene did not ground/refresh suggestion")
	}
	if _, err := testPool.Exec(ctx, `UPDATE world_facts SET player_known=false WHERE campaign_id=$1`, c.ID); err != nil {
		t.Fatal(err)
	}
	second := request(testUserID)
	var after game.ActionSuggestions
	if err := json.Unmarshal(second.Body.Bytes(), &after); err != nil {
		t.Fatal(err)
	}
	if before.Version == after.Version || strings.Contains(second.Body.String(), "Blue seal") {
		t.Fatal("known state change did not refresh candidates")
	}
	if _, err := q.UpdatePlayerCurrentHP(ctx, statedb.UpdatePlayerCurrentHPParams{ID: pc.ID, Hp: 0}); err != nil {
		t.Fatal(err)
	}
	var unavailable game.ActionSuggestions
	rec := request(testUserID)
	if err := json.Unmarshal(rec.Body.Bytes(), &unavailable); err != nil || len(unavailable.Candidates) != 0 {
		t.Fatal("unavailable actor suggested actions")
	}
	var writes int
	if err := testPool.QueryRow(ctx, `SELECT (SELECT count(*) FROM session_logs WHERE campaign_id=$1)+(SELECT count(*) FROM turn_attempts WHERE campaign_id=$1)`, c.ID).Scan(&writes); err != nil || writes != 0 {
		t.Fatalf("assistance wrote history=%d %v", writes, err)
	}
}
