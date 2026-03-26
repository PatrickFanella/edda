package game

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
	"github.com/PatrickFanella/game-master/internal/tools"
)

var _ tools.LanguageStore = (*worldService)(nil)
var _ tools.MemoryStore = (*worldService)(nil)

func TestWorldServiceCreateLanguage(t *testing.T) {
	q := newMockQuerier()
	langID := uuid.New()
	campaignID := uuid.New()
	factionID := uuid.New()

	q.createLanguageResult = statedb.Language{ID: dbutil.ToPgtype(langID)}
	svc := NewWorldService(q)

	got, err := svc.CreateLanguage(context.Background(), tools.CreateLanguageParams{
		CampaignID:         campaignID,
		Name:               "Elvish",
		Description:        "An ancient tongue.",
		SpokenByFactionIDs: []uuid.UUID{factionID},
	})
	if err != nil {
		t.Fatalf("CreateLanguage() error = %v", err)
	}
	if got != langID {
		t.Fatalf("returned ID = %s, want %s", got, langID)
	}
	if q.lastCreateLanguageParams == nil {
		t.Fatal("expected CreateLanguage to be called")
	}
	if q.lastCreateLanguageParams.Name != "Elvish" {
		t.Fatalf("name = %q, want Elvish", q.lastCreateLanguageParams.Name)
	}
}

func TestWorldServiceFactionBelongsToCampaignTrue(t *testing.T) {
	q := newMockQuerier()
	factionID := uuid.New()
	campaignID := uuid.New()

	q.faction = statedb.Faction{
		ID:         dbutil.ToPgtype(factionID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	svc := NewWorldService(q)

	ok, err := svc.FactionBelongsToCampaign(context.Background(), factionID, campaignID)
	if err != nil {
		t.Fatalf("FactionBelongsToCampaign() error = %v", err)
	}
	if !ok {
		t.Fatal("expected faction to belong to campaign")
	}
}

func TestWorldServiceFactionBelongsToCampaignFalse(t *testing.T) {
	q := newMockQuerier()
	factionID := uuid.New()
	campaignID := uuid.New()
	otherCampaignID := uuid.New()

	q.faction = statedb.Faction{
		ID:         dbutil.ToPgtype(factionID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	svc := NewWorldService(q)

	ok, err := svc.FactionBelongsToCampaign(context.Background(), factionID, otherCampaignID)
	if err != nil {
		t.Fatalf("FactionBelongsToCampaign() error = %v", err)
	}
	if ok {
		t.Fatal("expected faction to NOT belong to different campaign")
	}
}

func TestWorldServiceCultureBelongsToCampaignTrue(t *testing.T) {
	q := newMockQuerier()
	cultureID := uuid.New()
	campaignID := uuid.New()

	q.culture = statedb.Culture{
		ID:         dbutil.ToPgtype(cultureID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	svc := NewWorldService(q)

	ok, err := svc.CultureBelongsToCampaign(context.Background(), cultureID, campaignID)
	if err != nil {
		t.Fatalf("CultureBelongsToCampaign() error = %v", err)
	}
	if !ok {
		t.Fatal("expected culture to belong to campaign")
	}
}

func TestWorldServiceCultureBelongsToCampaignFalse(t *testing.T) {
	q := newMockQuerier()
	cultureID := uuid.New()
	campaignID := uuid.New()

	q.culture = statedb.Culture{
		ID:         dbutil.ToPgtype(cultureID),
		CampaignID: dbutil.ToPgtype(campaignID),
	}
	svc := NewWorldService(q)

	ok, err := svc.CultureBelongsToCampaign(context.Background(), cultureID, uuid.New())
	if err != nil {
		t.Fatalf("CultureBelongsToCampaign() error = %v", err)
	}
	if ok {
		t.Fatal("expected culture to NOT belong to different campaign")
	}
}

func TestWorldServiceCreateMemory(t *testing.T) {
	q := newMockQuerier()
	campaignID := uuid.New()
	svc := NewWorldService(q)

	embedding := []float32{0.1, 0.2, 0.3}
	err := svc.CreateMemory(context.Background(), tools.CreateMemoryParams{
		CampaignID: campaignID,
		Content:    "The old king vanished twenty years ago.",
		Embedding:  embedding,
		MemoryType: "world_fact",
		Metadata:   []byte(`{"source":"lore"}`),
	})
	if err != nil {
		t.Fatalf("CreateMemory() error = %v", err)
	}
	if q.lastCreateMemoryParams == nil {
		t.Fatal("expected CreateMemory to be called")
	}
	if q.lastCreateMemoryParams.Content != "The old king vanished twenty years ago." {
		t.Fatalf("content = %q", q.lastCreateMemoryParams.Content)
	}
	gotVec := q.lastCreateMemoryParams.Embedding.Slice()
	if len(gotVec) != len(embedding) {
		t.Fatalf("embedding length = %d, want %d", len(gotVec), len(embedding))
	}
	for i := range embedding {
		if gotVec[i] != embedding[i] {
			t.Fatalf("embedding[%d] = %f, want %f", i, gotVec[i], embedding[i])
		}
	}
}
