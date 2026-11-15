//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http/httptest"
	"os"
	"testing"
)

func TestIntegrationCharacterProjection(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, version := range []string{"legacy", "srd-5.2.1", "unsupported"} {
		rulesVersion := version
		build := stats
		if version == "unsupported" {
			rulesVersion = "srd-5.2.1"
			build = []byte(`{"srd521":{"class":"wizard"}}`)
		}
		c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Projection", RulesVersion: rulesVersion, Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
		if err != nil {
			t.Fatal(err)
		}
		_, err = q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Fixture fighter", Stats: build, Hp: 9, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
		if err != nil {
			t.Fatal(err)
		}
		h := &handlers.CharacterHandlers{Queries: q, Logger: log.New(io.Discard)}
		r := chi.NewRouter()
		r.Get("/campaigns/{id}/character", h.GetCharacter)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest("GET", "/campaigns/"+dbutil.FromPgtype(c.ID).String()+"/character", nil))
		if rec.Code != 200 {
			t.Fatalf("response %d %s", rec.Code, rec.Body.String())
		}
		var response api.CharacterResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatal(err)
		}
		if response.RulesVersion != rulesVersion || response.HP != 9 {
			t.Fatalf("bad adapter %+v", response)
		}
		if version == "unsupported" && (response.Sheet != nil || response.SheetUnavailable == "") {
			t.Fatalf("unsupported projection = %+v", response)
		}
		if version == "legacy" && response.Sheet != nil {
			t.Fatal("legacy converted")
		}
		if version == "srd-5.2.1" && (response.Sheet == nil || response.Sheet.Skills["athletics"].Value != 5 || response.Sheet.Progression.NextLevelXP != 300) {
			t.Fatalf("projection missing %+v", response)
		}
	}
}
