package game

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"git.subcult.tv/subculture-collective/edda/internal/dbutil"
	"git.subcult.tv/subculture-collective/edda/internal/domain"
	statedb "git.subcult.tv/subculture-collective/edda/internal/state/sqlc"
)

func TestQuestMutationUpdateAppendsAfterSparseObjectiveOrder(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	questID := uuid.New()
	q := newMockQuerier()
	q.quests = []statedb.Quest{questRow(questID, campaignID, uuid.Nil, "Recover Relic", domain.QuestStatusActive)}
	q.objectivesByQuest[dbutil.ToPgtype(questID).Bytes] = []statedb.QuestObjective{
		objectiveRow(uuid.New(), questID, "Find map", false, 5),
	}

	result, err := NewWorldService(q).UpdateQuest(context.Background(), domain.UpdateQuestMutationCommand{
		CampaignID:  campaignID,
		QuestID:     questID,
		Objectives:  []string{"Enter ruins"},
		Description: stringPtr("New briefing"),
	})
	if err != nil {
		t.Fatalf("UpdateQuest: %v", err)
	}
	if q.lastCreateObjectiveParams == nil || q.lastCreateObjectiveParams.OrderIndex != 6 {
		t.Fatalf("created objective order = %#v, want 6", q.lastCreateObjectiveParams)
	}
	if q.lastUpdateQuestParams == nil || q.lastUpdateQuestParams.Description.String != "New briefing" {
		t.Fatalf("description update = %#v, want New briefing", q.lastUpdateQuestParams)
	}
	if len(result.AddedObjectives) != 1 || result.AddedObjectives[0].OrderIndex != 6 {
		t.Fatalf("added objectives = %#v, want one at order 6", result.AddedObjectives)
	}
	if !strings.Contains(result.Narrative, "description updated") || !strings.Contains(result.Narrative, "1 new objective(s) added") {
		t.Fatalf("narrative = %q, want description/objective clauses", result.Narrative)
	}
}

func TestQuestMutationUpdateCascadesCompletedAndFailedOnly(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	parentID := uuid.New()
	activeChildID := uuid.New()
	doneChildID := uuid.New()
	grandchildID := uuid.New()
	q := newMockQuerier()
	q.quests = []statedb.Quest{
		questRow(parentID, campaignID, uuid.Nil, "Main", domain.QuestStatusActive),
		questRow(activeChildID, campaignID, parentID, "Child", domain.QuestStatusActive),
		questRow(doneChildID, campaignID, parentID, "Done Child", domain.QuestStatusCompleted),
		questRow(grandchildID, campaignID, activeChildID, "Grandchild", domain.QuestStatusActive),
	}

	completed := domain.QuestStatusCompleted
	result, err := NewWorldService(q).UpdateQuest(context.Background(), domain.UpdateQuestMutationCommand{
		CampaignID: campaignID,
		QuestID:    parentID,
		Status:     &completed,
	})
	if err != nil {
		t.Fatalf("UpdateQuest completed: %v", err)
	}
	if len(result.CascadedQuests) != 2 {
		t.Fatalf("cascaded = %#v, want active child and grandchild only", result.CascadedQuests)
	}
	for _, cascaded := range result.CascadedQuests {
		if cascaded.NewStatus != domain.QuestStatusAbandoned {
			t.Fatalf("cascade status = %s, want abandoned", cascaded.NewStatus)
		}
	}

	q = newMockQuerier()
	q.quests = []statedb.Quest{
		questRow(parentID, campaignID, uuid.Nil, "Main", domain.QuestStatusActive),
		questRow(activeChildID, campaignID, parentID, "Child", domain.QuestStatusActive),
	}
	abandoned := domain.QuestStatusAbandoned
	result, err = NewWorldService(q).UpdateQuest(context.Background(), domain.UpdateQuestMutationCommand{
		CampaignID: campaignID,
		QuestID:    parentID,
		Status:     &abandoned,
	})
	if err != nil {
		t.Fatalf("UpdateQuest abandoned: %v", err)
	}
	if len(result.CascadedQuests) != 0 {
		t.Fatalf("abandoned cascaded = %#v, want no cascade", result.CascadedQuests)
	}
}

func TestQuestMutationCompleteObjectivePreservesSelectionAndAlreadyComplete(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	questID := uuid.New()
	firstID := uuid.New()
	secondID := uuid.New()
	q := newMockQuerier()
	q.quests = []statedb.Quest{questRow(questID, campaignID, uuid.Nil, "Seal Gate", domain.QuestStatusActive)}
	q.objectivesByQuest[dbutil.ToPgtype(questID).Bytes] = []statedb.QuestObjective{
		objectiveRow(firstID, questID, "Find_the gate", true, 0),
		objectiveRow(secondID, questID, "Defeat guardian", false, 1),
	}
	desc := "find the   gate"

	result, err := NewWorldService(q).CompleteObjective(context.Background(), domain.CompleteObjectiveMutationCommand{
		CampaignID:           campaignID,
		QuestID:              questID,
		ObjectiveDescription: &desc,
	})
	if err != nil {
		t.Fatalf("CompleteObjective: %v", err)
	}
	if len(q.completeObjectiveCalls) != 0 {
		t.Fatalf("complete calls = %d, want 0 for already-complete objective", len(q.completeObjectiveCalls))
	}
	if !result.ObjectiveWasComplete || !strings.Contains(result.Narrative, "was already complete") {
		t.Fatalf("already complete result = %#v narrative=%q", result, result.Narrative)
	}
}

func TestQuestMutationCompleteObjectiveAutoCompleteCascadesAndNotReady(t *testing.T) {
	t.Parallel()

	campaignID := uuid.New()
	questID := uuid.New()
	objectiveID := uuid.New()
	childID := uuid.New()
	q := newMockQuerier()
	q.quests = []statedb.Quest{
		questRow(questID, campaignID, uuid.Nil, "Seal Gate", domain.QuestStatusActive),
		questRow(childID, campaignID, questID, "Child", domain.QuestStatusActive),
	}
	q.objectivesByQuest[dbutil.ToPgtype(questID).Bytes] = []statedb.QuestObjective{
		objectiveRow(objectiveID, questID, "Defeat guardian", false, 0),
	}

	result, err := NewWorldService(q).CompleteObjective(context.Background(), domain.CompleteObjectiveMutationCommand{
		CampaignID:        campaignID,
		QuestID:           questID,
		ObjectiveID:       &objectiveID,
		AutoCompleteQuest: true,
	})
	if err != nil {
		t.Fatalf("CompleteObjective: %v", err)
	}
	if !result.QuestAutoCompleted {
		t.Fatalf("QuestAutoCompleted = false, want true")
	}
	if result.QuestReadyForCompletion {
		t.Fatalf("QuestReadyForCompletion = true, want false after auto-complete")
	}
	if len(q.updateQuestStatusCalls) != 2 {
		t.Fatalf("status calls = %#v, want quest complete plus child cascade", q.updateQuestStatusCalls)
	}
	if q.updateQuestStatusCalls[1].Status != string(domain.QuestStatusAbandoned) {
		t.Fatalf("child status = %q, want abandoned", q.updateQuestStatusCalls[1].Status)
	}
}

func TestQuestMutationRejectsCrossCampaignQuest(t *testing.T) {
	t.Parallel()

	questID := uuid.New()
	q := newMockQuerier()
	q.quests = []statedb.Quest{questRow(questID, uuid.New(), uuid.Nil, "Other Campaign", domain.QuestStatusActive)}
	status := domain.QuestStatusCompleted

	_, err := NewWorldService(q).UpdateQuest(context.Background(), domain.UpdateQuestMutationCommand{
		CampaignID: uuid.New(),
		QuestID:    questID,
		Status:     &status,
	})
	if err == nil || !strings.Contains(err.Error(), "quest_id does not reference an existing quest") {
		t.Fatalf("err = %v, want not found for cross-campaign quest", err)
	}
	if len(q.updateQuestStatusCalls) != 0 {
		t.Fatalf("status calls = %#v, want no mutation", q.updateQuestStatusCalls)
	}
}

func questRow(id, campaignID, parentID uuid.UUID, title string, status domain.QuestStatus) statedb.Quest {
	quest := statedb.Quest{
		ID:          dbutil.ToPgtype(id),
		CampaignID:  dbutil.ToPgtype(campaignID),
		Title:       title,
		Description: pgtype.Text{String: "Description", Valid: true},
		QuestType:   string(domain.QuestTypeLongTerm),
		Status:      string(status),
	}
	if parentID != uuid.Nil {
		quest.ParentQuestID = dbutil.ToPgtype(parentID)
	}
	return quest
}

func objectiveRow(id, questID uuid.UUID, description string, completed bool, orderIndex int32) statedb.QuestObjective {
	return statedb.QuestObjective{
		ID:          dbutil.ToPgtype(id),
		QuestID:     dbutil.ToPgtype(questID),
		Description: description,
		Completed:   completed,
		OrderIndex:  orderIndex,
	}
}

func stringPtr(value string) *string { return &value }
