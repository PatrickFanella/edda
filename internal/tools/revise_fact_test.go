package tools

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubReviseFactStore struct {
	result *ReviseWorldFactResult
	err    error
	last   ReviseWorldFactCommand
}

func (s *stubReviseFactStore) ReviseWorldFact(_ context.Context, cmd ReviseWorldFactCommand) (*ReviseWorldFactResult, error) {
	s.last = cmd
	if s.err != nil {
		return nil, s.err
	}
	return s.result, nil
}

type stubReviseMemoryStore struct {
	called bool
	err    error
}

func (s *stubReviseMemoryStore) CreateMemory(_ context.Context, params CreateMemoryParams) error {
	s.called = true
	if s.err != nil {
		return s.err
	}
	return nil
}

type stubReviseEmbedder struct{ err error }

func (s *stubReviseEmbedder) Embed(_ context.Context, input string) ([]float32, error) {
	if s.err != nil {
		return nil, s.err
	}
	return []float32{1, 2}, nil
}

func TestReviseFactHandleSuccessIncludesCampaignID(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	newFactID := uuid.New()
	store := &stubReviseFactStore{result: &ReviseWorldFactResult{NewFact: domain.WorldFact{ID: newFactID, CampaignID: campaignID, Fact: "new", Category: "history"}}}
	h := NewReviseFactHandler(store, nil, nil)
	_, err := h.Handle(context.Background(), map[string]any{"campaign_id": campaignID.String(), "fact_id": factID.String(), "new_fact": "new"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if store.last.CampaignID != campaignID || store.last.FactID != factID || store.last.NewFact != "new" {
		t.Fatalf("command = %+v", store.last)
	}
}

func TestReviseFactHandleUsesCampaignIDFromContext(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	newFactID := uuid.New()
	store := &stubReviseFactStore{result: &ReviseWorldFactResult{NewFact: domain.WorldFact{ID: newFactID, CampaignID: campaignID, Fact: "new", Category: "history"}}}
	h := NewReviseFactHandler(store, nil, nil)
	ctx := WithCurrentCampaignID(context.Background(), campaignID)
	_, err := h.Handle(ctx, map[string]any{"fact_id": factID.String(), "new_fact": "new"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if store.last.CampaignID != campaignID {
		t.Fatalf("campaign_id = %s, want %s", store.last.CampaignID, campaignID)
	}
}

func TestReviseFactHandleRejectsCampaignIDMismatch(t *testing.T) {
	store := &stubReviseFactStore{}
	h := NewReviseFactHandler(store, nil, nil)
	ctx := WithCurrentCampaignID(context.Background(), uuid.New())
	_, err := h.Handle(ctx, map[string]any{"campaign_id": uuid.New().String(), "fact_id": uuid.New().String(), "new_fact": "new"})
	if err == nil {
		t.Fatal("expected campaign mismatch error")
	}
}

func TestReviseFactHandleAlreadySuperseded(t *testing.T) {
	store := &stubReviseFactStore{err: ErrWorldFactSuperseded}
	h := NewReviseFactHandler(store, nil, nil)
	_, err := h.Handle(context.Background(), map[string]any{"campaign_id": uuid.New().String(), "fact_id": uuid.New().String(), "new_fact": "new"})
	if err == nil || err.Error() == ErrWorldFactSuperseded.Error() {
		t.Fatalf("expected wrapped user-facing error, got %v", err)
	}
}

func TestReviseFactHandleEmbeddingFailureDoesNotFailMutation(t *testing.T) {
	campaignID := uuid.New()
	factID := uuid.New()
	newFactID := uuid.New()
	store := &stubReviseFactStore{result: &ReviseWorldFactResult{NewFact: domain.WorldFact{ID: newFactID, CampaignID: campaignID, Fact: "new", Category: "history"}}}
	h := NewReviseFactHandler(store, &stubReviseMemoryStore{}, &stubReviseEmbedder{err: errors.New("embed failed")})
	result, err := h.Handle(context.Background(), map[string]any{"campaign_id": campaignID.String(), "fact_id": factID.String(), "new_fact": "new"})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !result.Success || result.Data["memory_warning"] == nil {
		t.Fatalf("expected success with warning, got %+v", result)
	}
}

func TestReviseFactHandleNilStore(t *testing.T) {
	h := &ReviseFactHandler{}
	_, err := h.Handle(context.Background(), map[string]any{"campaign_id": uuid.New().String(), "fact_id": uuid.New().String(), "new_fact": "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}
