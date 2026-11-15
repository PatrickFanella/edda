//go:build integration

package handlers_test

import (
	"context"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/auth"
	"git.subcult.tv/subculture-collective/edda/internal/config"
	"git.subcult.tv/subculture-collective/edda/internal/db"
	"git.subcult.tv/subculture-collective/edda/internal/journal"
	"git.subcult.tv/subculture-collective/edda/internal/saves"
	"git.subcult.tv/subculture-collective/edda/internal/server"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

func TestIntegrationOwnershipNestedResources(t *testing.T) {
	ctx := context.Background()
	q := statedb.New(testPool)
	other, err := q.CreateUser(ctx, "ownership-"+uuid.NewString())
	if err != nil {
		t.Fatal(err)
	}
	a, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Owner A", Status: "active", CreatedBy: db.ToPgUUID(testUserID), Themes: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	b, err := q.CreateCampaign(ctx, statedb.CreateCampaignParams{Name: "Owner B", Status: "active", CreatedBy: other.ID, Themes: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	quest, err := q.CreateQuest(ctx, statedb.CreateQuestParams{CampaignID: b.ID, Title: "Private quest", QuestType: "short_term", Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	note, err := q.CreateQuestNote(ctx, statedb.CreateQuestNoteParams{QuestID: quest.ID, Content: "Private note"})
	if err != nil {
		t.Fatal(err)
	}
	js := journal.NewStore(testPool)
	entry, err := js.CreateEntry(ctx, db.FromPgUUID(b.ID), "Private journal", "Secret text")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{}
	cfg.Server.JWTSecret = "integration-ownership"
	router := server.NewRouterWithProvider(log.New(io.Discard), &stubIntegrationEngine{}, q, nil, testPool, testUserID, cfg, saves.NewStore(testPool))
	request := func(user uuid.UUID, method, path, body string, want int) {
		t.Helper()
		token, err := auth.GenerateToken(user, cfg.Server.JWTSecret, time.Hour)
		if err != nil {
			t.Fatal(err)
		}
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != want {
			t.Fatalf("%s %s: %d %s, want %d", method, path, rec.Code, rec.Body.String(), want)
		}
	}
	baseA := "/api/v1/campaigns/" + db.FromPgUUID(a.ID).String()
	baseB := "/api/v1/campaigns/" + db.FromPgUUID(b.ID).String()
	questPath := "/quests/" + db.FromPgUUID(quest.ID).String()
	notePath := questPath + "/notes/" + db.FromPgUUID(note.ID).String()
	// A cannot access B's children through A's otherwise authorized parent.
	for _, tc := range []struct{ method, path string }{{"GET", questPath + "/notes"}, {"POST", questPath + "/notes"}, {"DELETE", notePath}, {"GET", questPath + "/history"}} {
		request(testUserID, tc.method, baseA+tc.path, `{"content":"changed"}`, 404)
	}
	request(testUserID, "DELETE", baseB, "", 404)
	request(testUserID, "DELETE", baseA+"/journal/entries/"+entry.ID.String(), "", 204)
	entries, err := js.ListEntries(ctx, db.FromPgUUID(b.ID))
	if err != nil || len(entries) != 1 {
		t.Fatalf("foreign journal changed: %v %v", entries, err)
	}
	notes, err := q.ListQuestNotes(ctx, quest.ID)
	if err != nil || len(notes) != 1 || notes[0].Content != "Private note" {
		t.Fatalf("foreign notes changed: %v %v", notes, err)
	}
	// B can still use the same routes for B's own campaign and children.
	ownerB := db.FromPgUUID(other.ID)
	for _, path := range []string{"", questPath + "/notes", questPath + "/history", "/saves", "/time", "/journal/entries", "/journal/summaries", "/export/json", "/export/transcript"} {
		request(ownerB, "GET", baseB+path, "", 200)
	}
	request(ownerB, "POST", baseB+questPath+"/notes", `{"content":"Owner note"}`, 201)
	request(ownerB, "DELETE", baseB+notePath, "", 204)
	request(ownerB, "DELETE", baseB+"/journal/entries/"+entry.ID.String(), "", 204)
	entries, err = js.ListEntries(ctx, db.FromPgUUID(b.ID))
	if err != nil || len(entries) != 0 {
		t.Fatalf("owner journal delete: %v %v", entries, err)
	}
	// A note ID cannot be deleted via a different quest, even in the same campaign.
	second, err := q.CreateQuest(ctx, statedb.CreateQuestParams{CampaignID: b.ID, Title: "Second", QuestType: "short_term", Status: "active", Description: pgtype.Text{}})
	if err != nil {
		t.Fatal(err)
	}
	secondNote, err := q.CreateQuestNote(ctx, statedb.CreateQuestNoteParams{QuestID: second.ID, Content: "Second note"})
	if err != nil {
		t.Fatal(err)
	}
	request(ownerB, "DELETE", baseB+questPath+"/notes/"+db.FromPgUUID(secondNote.ID).String(), "", 204)
	notes, err = q.ListQuestNotes(ctx, second.ID)
	if err != nil || len(notes) != 1 {
		t.Fatalf("cross-quest delete changed note: %v %v", notes, err)
	}
}
