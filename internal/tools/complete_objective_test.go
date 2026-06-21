package tools

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubCompleteObjectiveStore struct {
	result *domain.CompleteObjectiveMutationResult
	cmds   []domain.CompleteObjectiveMutationCommand
	err    error
}

func (s *stubCompleteObjectiveStore) CompleteObjective(_ context.Context, cmd domain.CompleteObjectiveMutationCommand) (*domain.CompleteObjectiveMutationResult, error) {
	s.cmds = append(s.cmds, cmd)
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func TestCompleteObjectiveRequiresCampaignContext(t *testing.T) {
	h := NewCompleteObjectiveHandler(&stubCompleteObjectiveStore{})
	_, err := h.Handle(context.Background(), map[string]any{"quest_id": uuid.New().String(), "objective_id": uuid.New().String()})
	if err == nil || err.Error() != "current campaign context is required" {
		t.Fatalf("err=%v", err)
	}
}

func TestCompleteObjectivePassesDomainCommand(t *testing.T) {
	campaignID := uuid.New()
	questID := uuid.New()
	objectiveID := uuid.New()
	desc := "find it"
	result := &domain.CompleteObjectiveMutationResult{Quest: domain.QuestSummary{ID: questID, CampaignID: campaignID, Title: "Quest", Status: domain.QuestStatusCompleted}, Objective: domain.QuestObjective{ID: objectiveID, Description: desc, Completed: true}, ObjectivesCompleted: 1, ObjectivesTotal: 1, Progress: "1/1", AllObjectivesComplete: true, QuestAutoCompleted: true, Narrative: "ok"}
	store := &stubCompleteObjectiveStore{result: result}
	h := NewCompleteObjectiveHandler(store)
	got, err := h.Handle(campaignCtx(campaignID), map[string]any{"quest_id": questID.String(), "objective_id": objectiveID.String(), "auto_complete_quest": true})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(store.cmds) != 1 {
		t.Fatalf("cmd count=%d", len(store.cmds))
	}
	if store.cmds[0].CampaignID != campaignID || store.cmds[0].QuestID != questID {
		t.Fatalf("ids mismatch: %#v", store.cmds[0])
	}
	if store.cmds[0].ObjectiveID == nil || *store.cmds[0].ObjectiveID != objectiveID {
		t.Fatalf("objective id mismatch: %#v", store.cmds[0].ObjectiveID)
	}
	if !store.cmds[0].AutoCompleteQuest {
		t.Fatalf("auto-complete flag not passed")
	}
	if got.Data["quest_id"] != questID.String() {
		t.Fatalf("quest_id=%v", got.Data["quest_id"])
	}
	if got.Narrative != "ok" {
		t.Fatalf("narrative=%q", got.Narrative)
	}
}
