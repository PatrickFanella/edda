package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

type stubLinkQuestEntityStore struct {
	locationsByID        map[[16]byte]statedb.Location
	questsByID           map[[16]byte]statedb.Quest
	npcsByID             map[[16]byte]statedb.Npc
	factionsByID         map[[16]byte]statedb.Faction
	playerCharactersByID map[[16]byte]statedb.PlayerCharacter
	itemsByID            map[[16]byte]statedb.Item

	createRelationshipCalls []statedb.CreateRelationshipParams
	relationshipsByEntity   []statedb.EntityRelationship
}

func (s *stubLinkQuestEntityStore) GetLocationByID(_ context.Context, id pgtype.UUID) (statedb.Location, error) {
	loc, ok := s.locationsByID[id.Bytes]
	if !ok {
		return statedb.Location{}, pgx.ErrNoRows
	}
	return loc, nil
}

func (s *stubLinkQuestEntityStore) GetQuestByID(_ context.Context, id pgtype.UUID) (statedb.Quest, error) {
	q, ok := s.questsByID[id.Bytes]
	if !ok {
		return statedb.Quest{}, pgx.ErrNoRows
	}
	return q, nil
}

func (s *stubLinkQuestEntityStore) GetNPCByID(_ context.Context, id pgtype.UUID) (statedb.Npc, error) {
	npc, ok := s.npcsByID[id.Bytes]
	if !ok {
		return statedb.Npc{}, pgx.ErrNoRows
	}
	return npc, nil
}

func (s *stubLinkQuestEntityStore) GetFactionByID(_ context.Context, id pgtype.UUID) (statedb.Faction, error) {
	f, ok := s.factionsByID[id.Bytes]
	if !ok {
		return statedb.Faction{}, pgx.ErrNoRows
	}
	return f, nil
}

func (s *stubLinkQuestEntityStore) GetPlayerCharacterByID(_ context.Context, id pgtype.UUID) (statedb.PlayerCharacter, error) {
	pc, ok := s.playerCharactersByID[id.Bytes]
	if !ok {
		return statedb.PlayerCharacter{}, pgx.ErrNoRows
	}
	return pc, nil
}

func (s *stubLinkQuestEntityStore) GetItemByID(_ context.Context, id pgtype.UUID) (statedb.Item, error) {
	item, ok := s.itemsByID[id.Bytes]
	if !ok {
		return statedb.Item{}, pgx.ErrNoRows
	}
	return item, nil
}

func (s *stubLinkQuestEntityStore) CreateRelationship(_ context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error) {
	s.createRelationshipCalls = append(s.createRelationshipCalls, arg)
	return statedb.EntityRelationship{
		ID:               dbutil.ToPgtype(uuid.New()),
		CampaignID:       arg.CampaignID,
		SourceEntityType: arg.SourceEntityType,
		SourceEntityID:   arg.SourceEntityID,
		TargetEntityType: arg.TargetEntityType,
		TargetEntityID:   arg.TargetEntityID,
		RelationshipType: arg.RelationshipType,
		Description:      arg.Description,
		Strength:         arg.Strength,
	}, nil
}

func (s *stubLinkQuestEntityStore) GetRelationshipsByEntity(_ context.Context, _ statedb.GetRelationshipsByEntityParams) ([]statedb.EntityRelationship, error) {
	return s.relationshipsByEntity, nil
}

func newLinkQuestEntityFixture() (
	campaignID, currentLocationID, questID, npcID uuid.UUID,
	store *stubLinkQuestEntityStore,
	ctx context.Context,
) {
	campaignID = uuid.New()
	currentLocationID = uuid.New()
	questID = uuid.New()
	npcID = uuid.New()

	store = &stubLinkQuestEntityStore{
		locationsByID: map[[16]byte]statedb.Location{
			dbutil.ToPgtype(currentLocationID).Bytes: {ID: dbutil.ToPgtype(currentLocationID), CampaignID: dbutil.ToPgtype(campaignID)},
		},
		questsByID: map[[16]byte]statedb.Quest{
			dbutil.ToPgtype(questID).Bytes: {
				ID:         dbutil.ToPgtype(questID),
				CampaignID: dbutil.ToPgtype(campaignID),
				Title:      "Retrieve the Amulet",
				QuestType:  "short_term",
				Status:     "active",
			},
		},
		npcsByID: map[[16]byte]statedb.Npc{
			dbutil.ToPgtype(npcID).Bytes: {ID: dbutil.ToPgtype(npcID), CampaignID: dbutil.ToPgtype(campaignID)},
		},
	}
	ctx = WithCurrentLocationID(context.Background(), currentLocationID)
	return
}

func TestLinkQuestEntity_Success(t *testing.T) {
	_, _, questID, npcID, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	result, err := h.Handle(ctx, map[string]any{
		"quest_id":         questID.String(),
		"entity_type":      "npc",
		"entity_id":        npcID.String(),
		"link_description": "The NPC is the quest giver.",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if len(store.createRelationshipCalls) != 1 {
		t.Fatalf("CreateRelationship call count = %d, want 1", len(store.createRelationshipCalls))
	}
	call := store.createRelationshipCalls[0]
	if call.SourceEntityType != "quest" {
		t.Fatalf("source_entity_type = %q, want quest", call.SourceEntityType)
	}
	if call.TargetEntityType != "npc" {
		t.Fatalf("target_entity_type = %q, want npc", call.TargetEntityType)
	}
	if call.RelationshipType != "quest_related" {
		t.Fatalf("relationship_type = %q, want quest_related", call.RelationshipType)
	}
	if result.Data["relationship_type"] != "quest_related" {
		t.Fatalf("result relationship_type = %v, want quest_related", result.Data["relationship_type"])
	}
}

func TestLinkQuestEntity_QuestNotFound(t *testing.T) {
	_, _, _, npcID, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	_, err := h.Handle(ctx, map[string]any{
		"quest_id":         uuid.New().String(), // nonexistent
		"entity_type":      "npc",
		"entity_id":        npcID.String(),
		"link_description": "Irrelevant.",
	})
	if err == nil {
		t.Fatal("expected error for missing quest")
	}
	if !strings.Contains(err.Error(), "quest not found") {
		t.Fatalf("error = %q, want contains 'quest not found'", err.Error())
	}
}

func TestLinkQuestEntity_EntityNotFound(t *testing.T) {
	_, _, questID, _, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	_, err := h.Handle(ctx, map[string]any{
		"quest_id":         questID.String(),
		"entity_type":      "npc",
		"entity_id":        uuid.New().String(), // nonexistent NPC
		"link_description": "Irrelevant.",
	})
	if err == nil {
		t.Fatal("expected error for missing entity")
	}
	if !strings.Contains(err.Error(), "entity not found") {
		t.Fatalf("error = %q, want contains 'entity not found'", err.Error())
	}
}

func TestLinkQuestEntity_InvalidEntityType(t *testing.T) {
	_, _, questID, _, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	_, err := h.Handle(ctx, map[string]any{
		"quest_id":         questID.String(),
		"entity_type":      "dragon",
		"entity_id":        uuid.New().String(),
		"link_description": "Invalid type.",
	})
	if err == nil {
		t.Fatal("expected error for invalid entity type")
	}
	if !strings.Contains(err.Error(), "entity_type must be one of") {
		t.Fatalf("error = %q, want contains 'entity_type must be one of'", err.Error())
	}
}

func TestLinkQuestEntity_CustomLinkType(t *testing.T) {
	_, _, questID, npcID, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	result, err := h.Handle(ctx, map[string]any{
		"quest_id":         questID.String(),
		"entity_type":      "npc",
		"entity_id":        npcID.String(),
		"link_description": "The NPC is the quest target.",
		"link_type":        "quest_target",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(store.createRelationshipCalls) != 1 {
		t.Fatalf("CreateRelationship call count = %d, want 1", len(store.createRelationshipCalls))
	}
	if store.createRelationshipCalls[0].RelationshipType != "quest_target" {
		t.Fatalf("relationship_type = %q, want quest_target", store.createRelationshipCalls[0].RelationshipType)
	}
	if result.Data["relationship_type"] != "quest_target" {
		t.Fatalf("result relationship_type = %v, want quest_target", result.Data["relationship_type"])
	}
}

func TestLinkQuestEntity_MissingRequiredArgs(t *testing.T) {
	_, _, questID, npcID, store, ctx := newLinkQuestEntityFixture()
	h := NewLinkQuestEntityHandler(store)

	cases := []struct {
		name string
		args map[string]any
	}{
		{
			name: "missing quest_id",
			args: map[string]any{
				"entity_type":      "npc",
				"entity_id":        npcID.String(),
				"link_description": "Desc.",
			},
		},
		{
			name: "missing entity_id",
			args: map[string]any{
				"quest_id":         questID.String(),
				"entity_type":      "npc",
				"link_description": "Desc.",
			},
		},
		{
			name: "missing entity_type",
			args: map[string]any{
				"quest_id":         questID.String(),
				"entity_id":        npcID.String(),
				"link_description": "Desc.",
			},
		},
		{
			name: "missing link_description",
			args: map[string]any{
				"quest_id":    questID.String(),
				"entity_type": "npc",
				"entity_id":   npcID.String(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.Handle(ctx, tc.args)
			if err == nil {
				t.Fatal("expected error for missing required arg")
			}
		})
	}
}

func TestLinkQuestEntity_PlayerAliasNormalized(t *testing.T) {
	campaignID, currentLocationID, questID, _, store, ctx := newLinkQuestEntityFixture()
	playerID := uuid.New()
	store.playerCharactersByID = map[[16]byte]statedb.PlayerCharacter{
		dbutil.ToPgtype(playerID).Bytes: {ID: dbutil.ToPgtype(playerID), CampaignID: dbutil.ToPgtype(campaignID)},
	}
	_ = currentLocationID
	h := NewLinkQuestEntityHandler(store)

	result, err := h.Handle(ctx, map[string]any{
		"quest_id":         questID.String(),
		"entity_type":      "player",
		"entity_id":        playerID.String(),
		"link_description": "The player is involved.",
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if store.createRelationshipCalls[0].TargetEntityType != "player_character" {
		t.Fatalf("target_entity_type = %q, want player_character", store.createRelationshipCalls[0].TargetEntityType)
	}
	if result.Data["target_entity_type"] != "player_character" {
		t.Fatalf("result target_entity_type = %v, want player_character", result.Data["target_entity_type"])
	}
}
