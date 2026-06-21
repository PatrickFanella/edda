package tools

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubUpdateQuestStore struct {
	result *domain.UpdateQuestMutationResult
	cmds   []domain.UpdateQuestMutationCommand
	err    error
}

func (s *stubUpdateQuestStore) UpdateQuest(_ context.Context, cmd domain.UpdateQuestMutationCommand) (*domain.UpdateQuestMutationResult, error) {
	s.cmds = append(s.cmds, cmd)
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

func campaignCtx(id uuid.UUID) context.Context {
	return WithCurrentCampaignID(context.Background(), id)
}

func TestUpdateQuestRequiresCampaignContext(t *testing.T) {
	h := NewUpdateQuestHandler(&stubUpdateQuestStore{})
	_, err := h.Handle(context.Background(), map[string]any{"quest_id": uuid.New().String(), "status": "completed"})
	if err == nil || err.Error() != "current campaign context is required" {
		t.Fatalf("err=%v", err)
	}
}

func TestUpdateQuestPassesDomainCommand(t *testing.T) {
	campaignID := uuid.New()
	questID := uuid.New()
	desc := "New briefing"
	status := domain.QuestStatusCompleted
	store := &stubUpdateQuestStore{result: &domain.UpdateQuestMutationResult{Quest: domain.QuestSummary{ID: questID, CampaignID: campaignID, Title: "Quest", Description: desc, Status: status}, Narrative: "ok"}}
	h := NewUpdateQuestHandler(store)
	got, err := h.Handle(campaignCtx(campaignID), map[string]any{"quest_id": questID.String(), "status": "completed", "description_update": desc, "new_objectives": []any{"one"}})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if len(store.cmds) != 1 {
		t.Fatalf("cmd count=%d", len(store.cmds))
	}
	if store.cmds[0].CampaignID != campaignID || store.cmds[0].QuestID != questID {
		t.Fatalf("cmd ids mismatch: %#v", store.cmds[0])
	}
	if store.cmds[0].Status == nil || *store.cmds[0].Status != status {
		t.Fatalf("status mismatch: %#v", store.cmds[0].Status)
	}
	if store.cmds[0].Description == nil || *store.cmds[0].Description != desc {
		t.Fatalf("description mismatch: %#v", store.cmds[0].Description)
	}
	if got.Narrative != "ok" {
		t.Fatalf("narrative=%q", got.Narrative)
	}
	if got.Data["id"] != questID.String() {
		t.Fatalf("id=%v", got.Data["id"])
	}
}
