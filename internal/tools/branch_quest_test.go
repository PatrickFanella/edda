package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/PatrickFanella/game-master/internal/dbutil"
	"github.com/PatrickFanella/game-master/internal/domain"
	statedb "github.com/PatrickFanella/game-master/internal/state/sqlc"
)

type stubBranchQuestStore struct {
	locationsByID    map[[16]byte]statedb.Location
	questsByID       map[[16]byte]statedb.Quest
	createdQuest     statedb.Quest
	relationshipsByEntity []statedb.EntityRelationship

	getLocationErr         error
	getQuestErr            error
	createQuestErr         error
	createObjectiveErr     error
	createRelationshipErr  error
	updateQuestStatusErr   error
	listObjectivesErr      error
	getRelsByEntityErr     error

	createQuestCalls        []statedb.CreateQuestParams
	createObjectiveCalls    []statedb.CreateObjectiveParams
	createRelationshipCalls []statedb.CreateRelationshipParams
	updateQuestStatusCalls  []statedb.UpdateQuestStatusParams
}

func (s *stubBranchQuestStore) GetLocationByID(_ context.Context, id pgtype.UUID) (statedb.Location, error) {
	if s.getLocationErr != nil {
		return statedb.Location{}, s.getLocationErr
	}
	loc, ok := s.locationsByID[id.Bytes]
	if !ok {
		return statedb.Location{}, errors.New("location not found")
	}
	return loc, nil
}

func (s *stubBranchQuestStore) GetQuestByID(_ context.Context, id pgtype.UUID) (statedb.Quest, error) {
	if s.getQuestErr != nil {
		return statedb.Quest{}, s.getQuestErr
	}
	q, ok := s.questsByID[id.Bytes]
	if !ok {
		return statedb.Quest{}, pgx.ErrNoRows
	}
	return q, nil
}

func (s *stubBranchQuestStore) CreateQuest(_ context.Context, arg statedb.CreateQuestParams) (statedb.Quest, error) {
	if s.createQuestErr != nil {
		return statedb.Quest{}, s.createQuestErr
	}
	s.createQuestCalls = append(s.createQuestCalls, arg)
	if s.createdQuest.ID.Valid {
		return s.createdQuest, nil
	}
	return statedb.Quest{
		ID:          dbutil.ToPgtype(uuid.New()),
		CampaignID:  arg.CampaignID,
		Title:       arg.Title,
		Description: arg.Description,
		QuestType:   arg.QuestType,
		Status:      arg.Status,
	}, nil
}

func (s *stubBranchQuestStore) CreateObjective(_ context.Context, arg statedb.CreateObjectiveParams) (statedb.QuestObjective, error) {
	if s.createObjectiveErr != nil {
		return statedb.QuestObjective{}, s.createObjectiveErr
	}
	s.createObjectiveCalls = append(s.createObjectiveCalls, arg)
	return statedb.QuestObjective{
		ID:          dbutil.ToPgtype(uuid.New()),
		QuestID:     arg.QuestID,
		Description: arg.Description,
		Completed:   arg.Completed,
		OrderIndex:  arg.OrderIndex,
	}, nil
}

func (s *stubBranchQuestStore) CreateRelationship(_ context.Context, arg statedb.CreateRelationshipParams) (statedb.EntityRelationship, error) {
	if s.createRelationshipErr != nil {
		return statedb.EntityRelationship{}, s.createRelationshipErr
	}
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
	}, nil
}

func (s *stubBranchQuestStore) UpdateQuestStatus(_ context.Context, arg statedb.UpdateQuestStatusParams) (statedb.Quest, error) {
	if s.updateQuestStatusErr != nil {
		return statedb.Quest{}, s.updateQuestStatusErr
	}
	s.updateQuestStatusCalls = append(s.updateQuestStatusCalls, arg)
	q, ok := s.questsByID[arg.ID.Bytes]
	if !ok {
		return statedb.Quest{}, pgx.ErrNoRows
	}
	q.Status = arg.Status
	return q, nil
}

func (s *stubBranchQuestStore) ListObjectivesByQuest(_ context.Context, _ pgtype.UUID) ([]statedb.QuestObjective, error) {
	if s.listObjectivesErr != nil {
		return nil, s.listObjectivesErr
	}
	return nil, nil
}

func (s *stubBranchQuestStore) GetRelationshipsByEntity(_ context.Context, _ statedb.GetRelationshipsByEntityParams) ([]statedb.EntityRelationship, error) {
	if s.getRelsByEntityErr != nil {
		return nil, s.getRelsByEntityErr
	}
	return s.relationshipsByEntity, nil
}

// helpers

func branchQuestCtx(locationID uuid.UUID) context.Context {
	return WithCurrentLocationID(context.Background(), locationID)
}

func baseBranchArgs(originalQuestID uuid.UUID) map[string]any {
	return map[string]any{
		"original_quest_id":  originalQuestID.String(),
		"branch_title":       "New Path",
		"branch_description": "A divergent quest path.",
		"branch_objectives": []any{
			map[string]any{"description": "First objective", "order_index": float64(0)},
			map[string]any{"description": "Second objective", "order_index": float64(1)},
		},
		"reason": "Player chose diplomacy over combat",
	}
}

func baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID uuid.UUID) *stubBranchQuestStore {
	return &stubBranchQuestStore{
		locationsByID: map[[16]byte]statedb.Location{
			dbutil.ToPgtype(locationID).Bytes: {
				ID:         dbutil.ToPgtype(locationID),
				CampaignID: dbutil.ToPgtype(campaignID),
				Name:       "Tavern",
			},
		},
		questsByID: map[[16]byte]statedb.Quest{
			dbutil.ToPgtype(originalQuestID).Bytes: {
				ID:         dbutil.ToPgtype(originalQuestID),
				CampaignID: dbutil.ToPgtype(campaignID),
				Title:      "Original Quest",
				QuestType:  string(domain.QuestTypeMediumTerm),
				Status:     string(domain.QuestStatusActive),
			},
		},
		createdQuest: statedb.Quest{
			ID:          dbutil.ToPgtype(branchQuestID),
			CampaignID:  dbutil.ToPgtype(campaignID),
			Title:       "New Path",
			Description: pgtype.Text{String: "A divergent quest path.", Valid: true},
			QuestType:   string(domain.QuestTypeMediumTerm),
			Status:      string(domain.QuestStatusActive),
		},
	}
}

func TestBranchQuest_Success(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	originalQuestID := uuid.New()
	branchQuestID := uuid.New()

	store := baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID)
	handler := NewBranchQuestHandler(store)

	result, err := handler.Handle(branchQuestCtx(locationID), baseBranchArgs(originalQuestID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}
	if result.Data["id"] != branchQuestID.String() {
		t.Fatalf("branch quest id = %v, want %v", result.Data["id"], branchQuestID.String())
	}
	if result.Data["quest_type"] != string(domain.QuestTypeMediumTerm) {
		t.Fatalf("quest_type = %v, want medium_term (inherited)", result.Data["quest_type"])
	}

	// Verify quest was created with inherited type.
	if len(store.createQuestCalls) != 1 {
		t.Fatalf("create quest calls = %d, want 1", len(store.createQuestCalls))
	}
	if store.createQuestCalls[0].QuestType != string(domain.QuestTypeMediumTerm) {
		t.Fatalf("created quest type = %q, want medium_term", store.createQuestCalls[0].QuestType)
	}

	// 2 objectives created.
	if len(store.createObjectiveCalls) != 2 {
		t.Fatalf("create objective calls = %d, want 2", len(store.createObjectiveCalls))
	}

	// 1 relationship: quest_branch link (no original relationships to copy).
	if len(store.createRelationshipCalls) != 1 {
		t.Fatalf("create relationship calls = %d, want 1", len(store.createRelationshipCalls))
	}
	if store.createRelationshipCalls[0].RelationshipType != "quest_branch" {
		t.Fatalf("relationship type = %q, want quest_branch", store.createRelationshipCalls[0].RelationshipType)
	}

	// No original status update.
	if len(store.updateQuestStatusCalls) != 0 {
		t.Fatalf("update quest status calls = %d, want 0", len(store.updateQuestStatusCalls))
	}
	if result.Data["original_status_change"] != nil {
		t.Fatal("expected no original status change")
	}
}

func TestBranchQuest_WithOriginalStatusChange(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	originalQuestID := uuid.New()
	branchQuestID := uuid.New()

	store := baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID)
	handler := NewBranchQuestHandler(store)

	args := baseBranchArgs(originalQuestID)
	args["original_quest_status"] = "completed"

	result, err := handler.Handle(branchQuestCtx(locationID), args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}

	if len(store.updateQuestStatusCalls) != 1 {
		t.Fatalf("update quest status calls = %d, want 1", len(store.updateQuestStatusCalls))
	}
	if store.updateQuestStatusCalls[0].Status != string(domain.QuestStatusCompleted) {
		t.Fatalf("updated status = %q, want completed", store.updateQuestStatusCalls[0].Status)
	}

	statusChange, ok := result.Data["original_status_change"].(map[string]any)
	if !ok {
		t.Fatal("expected original_status_change map")
	}
	if statusChange["old_status"] != string(domain.QuestStatusActive) {
		t.Fatalf("old_status = %v, want active", statusChange["old_status"])
	}
	if statusChange["new_status"] != string(domain.QuestStatusCompleted) {
		t.Fatalf("new_status = %v, want completed", statusChange["new_status"])
	}
}

func TestBranchQuest_OriginalNotFound(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	missingQuestID := uuid.New()
	branchQuestID := uuid.New()

	store := baseBranchStore(locationID, campaignID, uuid.New(), branchQuestID)
	handler := NewBranchQuestHandler(store)

	_, err := handler.Handle(branchQuestCtx(locationID), baseBranchArgs(missingQuestID))
	if err == nil {
		t.Fatal("expected error for missing quest")
	}
	if !strings.Contains(err.Error(), "does not reference an existing quest") {
		t.Fatalf("error = %q, want 'does not reference an existing quest'", err.Error())
	}
}

func TestBranchQuest_InheritsRelationships(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	originalQuestID := uuid.New()
	branchQuestID := uuid.New()
	npcID := uuid.New()
	factionID := uuid.New()

	store := baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID)
	store.relationshipsByEntity = []statedb.EntityRelationship{
		{
			ID:               dbutil.ToPgtype(uuid.New()),
			CampaignID:       dbutil.ToPgtype(campaignID),
			SourceEntityType: "quest",
			SourceEntityID:   dbutil.ToPgtype(originalQuestID),
			TargetEntityType: "npc",
			TargetEntityID:   dbutil.ToPgtype(npcID),
			RelationshipType: "quest_related",
			Description:      pgtype.Text{String: "NPC link", Valid: true},
		},
		{
			ID:               dbutil.ToPgtype(uuid.New()),
			CampaignID:       dbutil.ToPgtype(campaignID),
			SourceEntityType: "faction",
			SourceEntityID:   dbutil.ToPgtype(factionID),
			TargetEntityType: "quest",
			TargetEntityID:   dbutil.ToPgtype(originalQuestID),
			RelationshipType: "quest_related",
			Description:      pgtype.Text{String: "Faction link", Valid: true},
		},
	}

	handler := NewBranchQuestHandler(store)
	result, err := handler.Handle(branchQuestCtx(locationID), baseBranchArgs(originalQuestID))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Fatal("expected success")
	}

	// 1 quest_branch + 2 copied = 3 total relationships.
	if len(store.createRelationshipCalls) != 3 {
		t.Fatalf("create relationship calls = %d, want 3", len(store.createRelationshipCalls))
	}

	// Verify first copied relationship has branch quest substituted for original.
	copiedRel1 := store.createRelationshipCalls[1]
	if copiedRel1.SourceEntityType != "quest" {
		t.Fatalf("copied rel source type = %q, want quest", copiedRel1.SourceEntityType)
	}
	if copiedRel1.SourceEntityID != dbutil.ToPgtype(branchQuestID) {
		t.Fatal("copied rel source ID should be branch quest, not original")
	}
	if copiedRel1.TargetEntityType != "npc" {
		t.Fatalf("copied rel target type = %q, want npc", copiedRel1.TargetEntityType)
	}

	// Verify second copied relationship has branch quest substituted for original in target.
	copiedRel2 := store.createRelationshipCalls[2]
	if copiedRel2.SourceEntityType != "faction" {
		t.Fatalf("copied rel2 source type = %q, want faction", copiedRel2.SourceEntityType)
	}
	if copiedRel2.TargetEntityID != dbutil.ToPgtype(branchQuestID) {
		t.Fatal("copied rel2 target ID should be branch quest, not original")
	}

	if result.Data["copied_relationships"] != 2 {
		t.Fatalf("copied_relationships = %v, want 2", result.Data["copied_relationships"])
	}
}

func TestBranchQuest_MissingRequiredArgs(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	originalQuestID := uuid.New()
	branchQuestID := uuid.New()

	store := baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID)
	handler := NewBranchQuestHandler(store)
	ctx := branchQuestCtx(locationID)

	tests := []struct {
		name   string
		remove string
	}{
		{"missing original_quest_id", "original_quest_id"},
		{"missing branch_title", "branch_title"},
		{"missing branch_description", "branch_description"},
		{"missing branch_objectives", "branch_objectives"},
		{"missing reason", "reason"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := baseBranchArgs(originalQuestID)
			delete(args, tc.remove)
			_, err := handler.Handle(ctx, args)
			if err == nil {
				t.Fatalf("expected error when %s is missing", tc.remove)
			}
		})
	}
}

func TestBranchQuest_InvalidOriginalStatus(t *testing.T) {
	locationID := uuid.New()
	campaignID := uuid.New()
	originalQuestID := uuid.New()
	branchQuestID := uuid.New()

	store := baseBranchStore(locationID, campaignID, originalQuestID, branchQuestID)
	handler := NewBranchQuestHandler(store)

	args := baseBranchArgs(originalQuestID)
	args["original_quest_status"] = "destroyed"

	_, err := handler.Handle(branchQuestCtx(locationID), args)
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	if !strings.Contains(err.Error(), "status must be one of") {
		t.Fatalf("error = %q, want status validation error", err.Error())
	}
}
