//go:build integration

package handlers_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestIntegrationReturnReminder(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	c, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Return fixture", Status: "active", CreatedBy: dbutil.ToPgtype(testUserID)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = testPool.Exec(ctx, `DELETE FROM campaigns WHERE id=$1`, c.ID) })
	cfg := config.Config{}
	cfg.Server.JWTSecret = "reminder-fixture"
	// No provider or engine is available: reminder reads must remain independent.
	router := server.NewRouterWithProvider(log.New(io.Discard), nil, q, nil, testPool, testUserID, cfg, nil)
	request := func(owner uuid.UUID) *httptest.ResponseRecorder {
		token, err := auth.GenerateToken(owner, cfg.Server.JWTSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest("GET", "/api/v1/campaigns/"+dbutil.FromPgtype(c.ID).String()+"/journal/reminder", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	if rec := request(uuid.New()); rec.Code != 404 {
		t.Fatalf("foreign owner=%d", rec.Code)
	}
	var empty journal.ReturnReminder
	rec := request(testUserID)
	if rec.Code != 200 {
		t.Fatalf("empty=%d %s", rec.Code, rec.Body)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &empty); err != nil || empty.Scene != nil || empty.Objective != nil || len(empty.KnownFacts) != 0 {
		t.Fatalf("empty=%+v %v", empty, err)
	}
	pc, err := q.CreatePlayerCharacter(ctx, statedb.CreatePlayerCharacterParams{CampaignID: c.ID, UserID: dbutil.ToPgtype(testUserID), Name: "Mira", Stats: []byte(`{}`), Abilities: []byte(`[]`), Status: "active", Hp: 5, MaxHp: 5, Level: 1})
	if err != nil {
		t.Fatal(err)
	}
	loc, err := q.CreateLocation(ctx, statedb.CreateLocationParams{CampaignID: c.ID, Name: "Hidden vault", Properties: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := q.UpdatePlayerLocation(ctx, statedb.UpdatePlayerLocationParams{ID: pc.ID, CurrentLocationID: loc.ID}); err != nil {
		t.Fatal(err)
	}
	for _, fact := range []struct {
		text  string
		known bool
	}{
		{"The blue seal is broken.", true}, {"Secret ruler is a dragon.", false}, {strings.Repeat("雪", 240), true},
	} {
		if _, err := q.CreateFact(ctx, statedb.CreateFactParams{CampaignID: c.ID, Fact: fact.text, Category: "clue", Source: "fixture", PlayerKnown: fact.known}); err != nil {
			t.Fatal(err)
		}
	}
	// A superseded clue must not revive forgotten state.
	if _, err := testPool.Exec(ctx, `INSERT INTO world_facts(campaign_id,fact,category,source,player_known,superseded_by) SELECT $1,'Obsolete clue','clue','fixture',true,id FROM world_facts WHERE campaign_id=$1 AND fact='The blue seal is broken.'`, c.ID); err != nil {
		t.Fatal(err)
	}
	for _, quest := range []struct{ title, status string }{{"Find the keeper", "active"}, {"Old secret objective", "completed"}, {"Abandoned secret objective", "abandoned"}} {
		if _, err := q.CreateQuest(ctx, statedb.CreateQuestParams{CampaignID: c.ID, Title: quest.title, Status: quest.status, QuestType: "short_term", Description: pgtype.Text{String: "Private design notes", Valid: true}}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := journal.NewStore(testPool).CreateSummary(ctx, dbutil.FromPgtype(c.ID), 1, 10, "Stale generated secret should never be reused."); err != nil {
		t.Fatal(err)
	}
	rec = request(testUserID)
	if rec.Code != 200 {
		t.Fatalf("reminder=%d %s", rec.Code, rec.Body)
	}
	for _, forbidden := range []string{"Hidden vault", "Secret ruler", "Obsolete clue", "Old secret", "Abandoned secret", "Private design", "Stale generated"} {
		if strings.Contains(rec.Body.String(), forbidden) {
			t.Fatalf("leaked %q in %s", forbidden, rec.Body)
		}
	}
	var reminder journal.ReturnReminder
	if err := json.Unmarshal(rec.Body.Bytes(), &reminder); err != nil {
		t.Fatal(err)
	}
	if reminder.Scene != nil || reminder.Objective == nil || reminder.Objective.Text != "Find the keeper" || len(reminder.KnownFacts) != 2 {
		t.Fatalf("reminder=%+v", reminder)
	}
	for _, fact := range reminder.KnownFacts {
		if len([]rune(fact.Text)) > 201 || !strings.HasPrefix(fact.SourceID, "known-fact:") {
			t.Fatalf("fact=%+v", fact)
		}
	}

	// Additional general knowledge cannot displace the relevant known clues,
	// and the response remains bounded at three facts.
	for _, text := range []string{"General fact A", "General fact B", "General fact C", "General fact D"} {
		if _, err := q.CreateFact(ctx, statedb.CreateFactParams{CampaignID: c.ID, Fact: text, Category: "history", Source: "fixture", PlayerKnown: true}); err != nil {
			t.Fatal(err)
		}
	}
	bounded := request(testUserID)
	var boundedReminder journal.ReturnReminder
	if err := json.Unmarshal(bounded.Body.Bytes(), &boundedReminder); err != nil {
		t.Fatal(err)
	}
	if bounded.Code != 200 || len(boundedReminder.KnownFacts) != 3 || !strings.Contains(bounded.Body.String(), "The blue seal is broken.") {
		t.Fatalf("bounded reminder=%d %+v", bounded.Code, boundedReminder)
	}
	if err := q.SetLocationPlayerVisited(ctx, loc.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := testPool.Exec(ctx, `UPDATE world_facts SET player_known=false WHERE campaign_id=$1`, c.ID); err != nil {
		t.Fatal(err)
	}
	rec = request(testUserID)
	if err := json.Unmarshal(rec.Body.Bytes(), &reminder); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || reminder.Scene == nil || reminder.Scene.Text != "Hidden vault" || len(reminder.KnownFacts) != 0 {
		t.Fatalf("refreshed=%d %+v", rec.Code, reminder)
	}
	var writes int
	if err := testPool.QueryRow(ctx, `SELECT (SELECT count(*) FROM session_logs WHERE campaign_id=$1)+(SELECT count(*) FROM turn_attempts WHERE campaign_id=$1)+(SELECT count(*) FROM save_points WHERE campaign_id=$1)`, c.ID).Scan(&writes); err != nil || writes != 0 {
		t.Fatalf("gameplay writes=%d %v", writes, err)
	}
}
