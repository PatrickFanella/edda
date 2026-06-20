package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"git.subcult.tv/subculture-collective/edda/internal/domain"
)

type stubRevealLocationStore struct {
	result *domain.RevealLocationResult
	err    error
	last   domain.RevealLocationCommand
}

func (s *stubRevealLocationStore) RevealLocation(_ context.Context, cmd domain.RevealLocationCommand) (*domain.RevealLocationResult, error) {
	s.last = cmd
	return s.result, s.err
}

func TestRevealLocationHandleSuccess(t *testing.T) {
	campaignID := uuid.New()
	locationID := uuid.New()
	h := NewRevealLocationHandler(&stubRevealLocationStore{result: &domain.RevealLocationResult{LocationName: "Ancient Gate"}})
	got, err := h.Handle(WithCurrentCampaignID(context.Background(), campaignID), map[string]any{"location_id": locationID.String()})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if got.Data["name"] != "Ancient Gate" {
		t.Fatalf("name = %v", got.Data["name"])
	}
}

func TestRevealLocationHandleMissingContext(t *testing.T) {
	h := NewRevealLocationHandler(&stubRevealLocationStore{result: &domain.RevealLocationResult{LocationName: "Ancient Gate"}})
	_, err := h.Handle(context.Background(), map[string]any{"location_id": uuid.New().String()})
	if err == nil || !strings.Contains(err.Error(), "current campaign id") {
		t.Fatalf("err = %v", err)
	}
}

func TestRevealLocationHandleCampaignScopeError(t *testing.T) {
	h := NewRevealLocationHandler(&stubRevealLocationStore{err: errors.New("location not found in campaign")})
	_, err := h.Handle(WithCurrentCampaignID(context.Background(), uuid.New()), map[string]any{"location_id": uuid.New().String()})
	if err == nil || !strings.Contains(err.Error(), "reveal location") {
		t.Fatalf("err = %v", err)
	}
}
