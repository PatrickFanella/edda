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

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestIntegrationActionPreviewReadOnly(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	stats, err := os.ReadFile("../rules/srd521/testdata/fighter1.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Preview", RulesVersion: "srd-5.2.1", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: stats, Hp: 9, MaxHp: 12, Level: 1, Status: "active", Abilities: []byte(`[]`)})
	if err != nil {
		t.Fatal(err)
	}
	h := &handlers.CharacterHandlers{Queries: q, Pool: testPool, Logger: log.New(io.Discard)}
	campaignH := &handlers.CampaignHandlers{Queries: q, Logger: log.New(io.Discard)}
	r := chi.NewRouter()
	r.Use(auth.NewNoOpMiddleware(testUserID).Authenticate)
	r.Route("/campaigns/{id}", func(r chi.Router) {
		r.Use(campaignH.RequireOwnership)
		r.Get("/character/action-preview", h.PreviewAction)
	})
	path := "/campaigns/" + dbutil.FromPgtype(c.ID).String() + "/character/action-preview?actor_id=" + dbutil.FromPgtype(pc.ID).String()
	request := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}
	snapshot := func() string {
		var raw string
		err := testPool.QueryRow(ctx, `SELECT jsonb_build_object('campaign',to_jsonb(c),'characters',(SELECT jsonb_agg(p) FROM player_characters p WHERE p.campaign_id=c.id),'clock',(SELECT jsonb_agg(t) FROM campaign_time t WHERE t.campaign_id=c.id),'logs',(SELECT jsonb_agg(l) FROM session_logs l WHERE l.campaign_id=c.id),'attempts',(SELECT jsonb_agg(a) FROM turn_attempts a WHERE a.campaign_id=c.id))::text FROM campaigns c WHERE c.id=$1`, c.ID).Scan(&raw)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	before := snapshot()
	var first api.ActionPreview
	for i := 0; i < 2; i++ {
		rec := request(path + "&kind=ability-check&skill=athletics")
		if rec.Code != 200 {
			t.Fatalf("preview %d: %s", rec.Code, rec.Body.String())
		}
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("preview cached")
		}
		var got api.ActionPreview
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if !got.Available || got.Modifier.Value != 5 || got.Intent.Ability != "strength" {
			t.Fatalf("bad preview %+v", got)
		}
		if i == 0 {
			first = got
		} else if !reflect.DeepEqual(first, got) {
			t.Fatal("repeat preview changed")
		}
	}
	if before != snapshot() {
		t.Fatal("preparation changed campaign, time, history or attempts")
	}
	for _, suffix := range []string{"&kind=attack&entry_id=secret-npc", "&kind=ability-check&ability=secret-ability"} {
		rec := request(path + suffix)
		var got api.ActionPreview
		if rec.Code != 200 || json.Unmarshal(rec.Body.Bytes(), &got) != nil || got.Available || got.Reason == "" {
			t.Fatalf("invalid action: %s", rec.Body.String())
		}
		if got.Modifier != nil || len(got.Targets) != 0 {
			t.Fatal("invalid action leaked projection")
		}
	}
	for _, suffix := range []string{"&kind=ability-check&ability=strength&difficulty=18", "&kind=ability-check&kind=saving-throw"} {
		if got := request(path + suffix); got.Code != 400 {
			t.Fatalf("unsafe parameters: %d", got.Code)
		}
	}
	foreign := strings.Replace(path, dbutil.FromPgtype(pc.ID).String(), uuid.NewString(), 1)
	if got := request(foreign + "&kind=ability-check&ability=strength"); got.Code != 404 {
		t.Fatalf("foreign actor %d", got.Code)
	}
	if _, err := q.UpdatePlayerCurrentHP(ctx, statedb.UpdatePlayerCurrentHPParams{ID: pc.ID, Hp: 8}); err != nil {
		t.Fatal(err)
	}
	rec := request(path + "&kind=ability-check&skill=athletics")
	var changed api.ActionPreview
	if err := json.Unmarshal(rec.Body.Bytes(), &changed); err != nil {
		t.Fatal(err)
	}
	if changed.ExpectedRevision == first.ExpectedRevision {
		t.Fatal("HP change did not invalidate preview")
	}
}
