//go:build integration

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/handlers"
	"git.subcult.tv/subculture-collective/edda/internal/rules/srd521"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"git.subcult.tv/subculture-collective/edda/internal/world"
	"git.subcult.tv/subculture-collective/edda/pkg/api"
	"github.com/charmbracelet/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestIntegrationInterviewAndGuidedCreationShareProjection(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	for _, class := range []string{"Fighter", "Cleric"} {
		t.Run(class, func(t *testing.T) {
			background := "Soldier"
			if class == "Cleric" {
				background = "Acolyte"
			}
			var previous *srd521.CharacterSheet
			for _, interview := range []bool{false, true} {
				p := &world.CharacterProfile{Name: "Mira", Concept: "Human " + class, Background: background}
				if interview {
					p.RulesChoices = &world.RulesChoices{Origin: "Human", Class: class, Background: background}
					p.Strengths = []string{"Narrative strength grants no extra ability"}
				}
				h := handlers.NewStartupHandlers(&creationProvider{t: t}, q, log.Default(), nil)
				router := chi.NewRouter()
				router.Use(auth.NewNoOpMiddleware(testUserID).Authenticate)
				router.Post("/world", h.BuildWorld)
				body, err := json.Marshal(api.WorldBuildRequest{Name: "Shared catalogue", Profile: creationInput(testUserID).Profile, CharacterProfile: p, RulesVersion: "srd-5.2.1"})
				if err != nil {
					t.Fatal(err)
				}
				rec := httptest.NewRecorder()
				router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/world", bytes.NewReader(body)))
				if rec.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
				}
				var response api.WorldBuildResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Fatal(err)
				}
				// Use the response ID to read the actual durable character.
				pcs, err := q.GetPlayerCharacterByCampaign(ctx, db.ToPgUUID(uuid.MustParse(response.Campaign.ID)))
				if err != nil || len(pcs) != 1 {
					t.Fatalf("pcs=%v err=%v", pcs, err)
				}
				t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, pcs[0].CampaignID) })
				pc := pcs[0]
				sheet, err := srd521.Project(pc.Stats, 1, int(pc.Hp), int(pc.MaxHp), 0)
				if err != nil {
					t.Fatal(err)
				}
				if previous != nil && !reflect.DeepEqual(previous, sheet) {
					t.Fatal("interview projection diverged from guided")
				}
				previous = sheet
				if bytes.Contains(pc.Abilities, []byte("Narrative strength")) {
					t.Fatal("narrative strength became mechanical ability")
				}
			}
		})
	}
}
