package tui

import (
	"context"
	"errors"
	"git.subcult.tv/subculture-collective/edda/internal/engine"
	"git.subcult.tv/subculture-collective/edda/internal/game"
	"git.subcult.tv/subculture-collective/edda/tui/narrative"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"strings"
	"testing"
)

type surgeMock struct {
	mockGameEngine
	requests []engine.PreparedAction
	fail     bool
}

func (m *surgeMock) PreviewActionSurge(context.Context, uuid.UUID) (*game.ActionPreview, error) {
	return &game.ActionPreview{Available: true, ExpectedRevision: "surge-rev", Intent: game.ActionIntent{ActorID: uuid.New(), Kind: "action-surge"}, Options: []string{"one additional Action except Magic"}}, nil
}
func (m *surgeMock) ConfirmAction(_ context.Context, _ uuid.UUID, r engine.PreparedAction) (*engine.TurnResult, error) {
	m.requests = append(m.requests, r)
	if m.fail {
		return nil, errors.New("lost response")
	}
	return &engine.TurnResult{ActionID: r.ActionID, Narrative: "Action Surge saved"}, nil
}
func TestTUIActionSurgePreviewConfirmationRetry(t *testing.T) {
	m := &surgeMock{fail: true}
	a := NewAppWithEngine(testCfg, testCampaign, context.Background(), m, nil)
	sized, _ := a.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	a = sized.(App)
	send := func(input string) {
		t.Helper()
		model, cmd := a.Update(narrative.SubmitMsg{Input: input})
		a = model.(App)
		if cmd != nil {
			model, _ = a.Update(cmd())
			a = model.(App)
		}
	}
	send("/confirm-action-surge")
	if len(m.requests) != 0 {
		t.Fatal("confirm without preview")
	}
	send("/action-surge")
	if a.surgeRequest == nil || len(m.requests) != 0 || !strings.Contains(a.narrativeView().View(), "Preview only") {
		t.Fatal("preview unusable")
	}
	send("/confirm-action-surge")
	if a.surgeRequest == nil || len(m.requests) != 1 {
		t.Fatal("lost identity")
	}
	m.fail = false
	send("/confirm-action-surge")
	if len(m.requests) != 2 || m.requests[0] != m.requests[1] || a.surgeRequest != nil || len(m.inputs) != 0 {
		t.Fatal("retry changed identity or used inference")
	}
	send("/action-surge")
	send("/cancel-action-surge")
	if a.surgeRequest != nil {
		t.Fatal("cancel failed")
	}
}
