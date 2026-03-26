package game

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

const innkeeperName = "Innkeeper Toma"

func TestNPCDialogueStoreGetNPCByID(t *testing.T) {
	q := newMockQuerier()
	npcID := uuid.New()
	campaignID := uuid.New()
	locationID := uuid.New()
	q.npcByID[npcID] = mockNPCRecord{
		npc: statedb.Npc{
			ID:         dbutil.ToPgtype(npcID),
			CampaignID: dbutil.ToPgtype(campaignID),
			Name:       innkeeperName,
			LocationID: dbutil.ToPgtype(locationID),
			Alive:      true,
		},
	}
	store := NewNPCService(q)

	got, err := store.GetNPCByID(context.Background(), npcID)
	if err != nil {
		t.Fatalf("GetNPCByID() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected NPC, got nil")
	}
	if got.ID != npcID {
		t.Fatalf("npc id = %s, want %s", got.ID, npcID)
	}
	if got.Name != innkeeperName {
		t.Fatalf("npc name = %q, want %s", got.Name, innkeeperName)
	}
	if got.LocationID == nil || *got.LocationID != locationID {
		t.Fatalf("npc location = %v, want %s", got.LocationID, locationID)
	}
}

func TestNPCDialogueStoreGetNPCByIDNotFound(t *testing.T) {
	q := newMockQuerier()
	store := NewNPCService(q)

	got, err := store.GetNPCByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetNPCByID() error = %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil npc, got %+v", got)
	}
}

func TestNPCDialogueStoreLogNPCDialoguePersistsSessionLog(t *testing.T) {
	q := newMockQuerier()
	store := NewNPCService(q)
	campaignID := uuid.New()
	locationID := uuid.New()
	npcID := uuid.New()
	q.recentSessionLogs = []statedb.SessionLog{{TurnNumber: 7}}

	err := store.LogNPCDialogue(context.Background(), tools.NPCDialogueLogEntry{
		CampaignID:        campaignID,
		LocationID:        locationID,
		NPCID:             npcID,
		Dialogue:          "Welcome, traveler.",
		FormattedDialogue: "Innkeeper Toma: Welcome, traveler.",
	})
	if err != nil {
		t.Fatalf("LogNPCDialogue() error = %v", err)
	}
	if q.lastSessionLog == nil {
		t.Fatal("expected CreateSessionLog to be called")
	}
	if q.lastSessionLog.TurnNumber != 8 {
		t.Fatalf("turn_number = %d, want 8", q.lastSessionLog.TurnNumber)
	}
	if q.lastSessionLog.InputType != string(domain.Narrative) {
		t.Fatalf("input_type = %q, want %q", q.lastSessionLog.InputType, domain.Narrative)
	}
	if q.lastSessionLog.PlayerInput != "Innkeeper Toma: Welcome, traveler." {
		t.Fatalf("player_input = %q", q.lastSessionLog.PlayerInput)
	}
	if q.lastSessionLog.LlmResponse != "Innkeeper Toma: Welcome, traveler." {
		t.Fatalf("llm_response = %q", q.lastSessionLog.LlmResponse)
	}
	if string(q.lastSessionLog.ToolCalls) != "[]" {
		t.Fatalf("tool_calls = %s, want []", q.lastSessionLog.ToolCalls)
	}
	if got := dbutil.FromPgtype(q.lastSessionLog.LocationID); got != locationID {
		t.Fatalf("location_id = %s, want %s", got, locationID)
	}
	if len(q.lastSessionLog.NpcsInvolved) != 1 || dbutil.FromPgtype(q.lastSessionLog.NpcsInvolved[0]) != npcID {
		t.Fatalf("npcs_involved = %+v, want [%s]", q.lastSessionLog.NpcsInvolved, npcID)
	}
}

func TestNPCDialogueStoreLogNPCDialogueErrors(t *testing.T) {
	t.Run("list recent session logs", func(t *testing.T) {
		q := newMockQuerier()
		q.listRecentSessionLogsErr = errors.New("query failed")
		store := NewNPCService(q)

		err := store.LogNPCDialogue(context.Background(), tools.NPCDialogueLogEntry{
			CampaignID:        uuid.New(),
			LocationID:        uuid.New(),
			NPCID:             uuid.New(),
			FormattedDialogue: "Guard: Halt!",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "list recent session logs: query failed") {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("create session log", func(t *testing.T) {
		q := newMockQuerier()
		q.createSessionLogErr = errors.New("insert failed")
		store := NewNPCService(q)

		err := store.LogNPCDialogue(context.Background(), tools.NPCDialogueLogEntry{
			CampaignID:        uuid.New(),
			LocationID:        uuid.New(),
			NPCID:             uuid.New(),
			FormattedDialogue: "Guard: Halt!",
		})
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "create session log: insert failed") {
			t.Fatalf("error = %v", err)
		}
	})
}

var _ tools.NPCDialogueStore = (*npcService)(nil)
var _ tools.UpdateNPCStore = (*npcService)(nil)

// ---------------------------------------------------------------------------
// UpdateNPC / LocationExistsInCampaign tests (merged from update_npc_store_test.go)
// ---------------------------------------------------------------------------

func TestNPCServiceLocationExistsInCampaign(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	otherCampaignID := uuid.New()

	q := newMockQuerier()
	q.location = statedb.Location{
		ID:         dbutil.ToPgtype(locationID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	store := NewNPCService(q)

	ok, err := store.LocationExistsInCampaign(context.Background(), locationID, campaignID)
	if err != nil {
		t.Fatalf("LocationExistsInCampaign() error = %v", err)
	}
	if !ok {
		t.Fatal("expected location to exist in campaign")
	}

	ok, err = store.LocationExistsInCampaign(context.Background(), locationID, otherCampaignID)
	if err != nil {
		t.Fatalf("LocationExistsInCampaign() error = %v", err)
	}
	if ok {
		t.Fatal("expected location to be rejected for different campaign")
	}
}

func TestStringToPgText(t *testing.T) {
	if got := stringToPgText(""); got.Valid {
		t.Fatalf("stringToPgText(\"\") valid = %v, want false", got.Valid)
	}
	got := stringToPgText("hello")
	if !got.Valid || got.String != "hello" {
		t.Fatalf("stringToPgText(\"hello\") = %+v", got)
	}
}

func TestNPCServiceUpdateNPCPreservesNullTextForEmptyStrings(t *testing.T) {
	npcID := uuid.New()
	q := newMockQuerier()
	q.updateNPCResult = statedb.Npc{
		ID:         dbutil.ToPgtype(npcID),
		CampaignID: dbutil.ToPgtype(uuid.New()),
		Name:       "Null Keeper",
	}
	store := NewNPCService(q)

	_, err := store.UpdateNPC(context.Background(), domain.NPC{
		ID:          npcID,
		CampaignID:  uuid.New(),
		Name:        "Null Keeper",
		Description: "",
		Personality: "",
		Disposition: 1,
		Alive:       true,
	})
	if err != nil {
		t.Fatalf("UpdateNPC() error = %v", err)
	}
	if q.lastUpdateNPCParams == nil {
		t.Fatal("expected UpdateNPC to be called")
	}
	if q.lastUpdateNPCParams.Description.Valid {
		t.Fatalf("description valid = %v, want false", q.lastUpdateNPCParams.Description.Valid)
	}
	if q.lastUpdateNPCParams.Personality.Valid {
		t.Fatalf("personality valid = %v, want false", q.lastUpdateNPCParams.Personality.Valid)
	}
}

func TestIntOrNullInt4(t *testing.T) {
	if got := intOrNullInt4(nil); got.Valid {
		t.Fatalf("intOrNullInt4(nil) valid = %v, want false", got.Valid)
	}
	v := 7
	got := intOrNullInt4(&v)
	if !got.Valid || got.Int32 != 7 {
		t.Fatalf("intOrNullInt4(&7) = %+v", got)
	}
}

func TestNPCServiceUpdateNPCPassesNullableFields(t *testing.T) {
	npcID := uuid.New()
	locationID := uuid.New()
	factionID := uuid.New()
	hp := 11

	q := newMockQuerier()
	q.updateNPCResult = statedb.Npc{
		ID:         dbutil.ToPgtype(npcID),
		CampaignID: dbutil.ToPgtype(uuid.New()),
		Name:       "Ranger",
	}
	store := NewNPCService(q)

	_, err := store.UpdateNPC(context.Background(), domain.NPC{
		ID:          npcID,
		CampaignID:  uuid.New(),
		Name:        "Ranger",
		Description: "alert",
		Personality: "calm",
		Disposition: 9,
		LocationID:  &locationID,
		FactionID:   &factionID,
		Alive:       true,
		HP:          &hp,
	})
	if err != nil {
		t.Fatalf("UpdateNPC() error = %v", err)
	}
	if q.lastUpdateNPCParams == nil {
		t.Fatal("expected UpdateNPC to be called")
	}
	if !q.lastUpdateNPCParams.LocationID.Valid {
		t.Fatal("expected valid location_id")
	}
	if !q.lastUpdateNPCParams.FactionID.Valid {
		t.Fatal("expected valid faction_id")
	}
	if !q.lastUpdateNPCParams.Hp.Valid || q.lastUpdateNPCParams.Hp.Int32 != int32(hp) {
		t.Fatalf("expected hp=%d valid, got %+v", hp, q.lastUpdateNPCParams.Hp)
	}
}

func TestNPCServiceLocationExistsInCampaignNotFound(t *testing.T) {
	q := newMockQuerier()
	store := NewNPCService(q)

	ok, err := store.LocationExistsInCampaign(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("LocationExistsInCampaign() error = %v", err)
	}
	if ok {
		t.Fatal("expected false when location does not exist")
	}
}

func TestStringToPgTextRoundTrip(t *testing.T) {
	value := "story"
	got := stringToPgText(value)
	if got != (pgtype.Text{String: value, Valid: true}) {
		t.Fatalf("unexpected pgtype.Text: %+v", got)
	}
}
