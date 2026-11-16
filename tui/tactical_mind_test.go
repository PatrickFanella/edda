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

type tacticalMock struct {
	mockGameEngine
	requests []engine.PreparedAction
	fail     bool
}

func (m *tacticalMock) PreviewTacticalMind(_ context.Context, _ uuid.UUID, choice string) (*game.ActionPreview, error) {
	return &game.ActionPreview{Available: true, ExpectedRevision: "tactical-rev", Intent: game.ActionIntent{ActorID: uuid.New(), Kind: "tactical-mind", EntryID: choice}, Options: []string{"one d10; Second Wind only on success"}}, nil
}
func (m *tacticalMock) PreviewEasyAthletics(context.Context, uuid.UUID) (*game.ActionPreview, error) {
	return &game.ActionPreview{Available: true, ExpectedRevision: "easy-rev", Intent: game.ActionIntent{ActorID: uuid.New(), Kind: "ability-check", EntryID: game.EasyAthleticsEntry}, Options: []string{"Edda bounded adjudication fixture: Easy Strength (Athletics), DC 10"}}, nil
}
func (m *tacticalMock) ConfirmAction(_ context.Context, _ uuid.UUID, r engine.PreparedAction) (*engine.TurnResult, error) {
	m.requests = append(m.requests, r)
	if m.fail {
		return nil, errors.New("lost response")
	}
	return &engine.TurnResult{ActionID: r.ActionID, Narrative: "Tactical Mind saved"}, nil
}
func TestTUITacticalMindPreviewConfirmationRetry(t *testing.T) {
	m := &tacticalMock{fail: true}
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
	send("/confirm-tactical-mind")
	if len(m.requests) != 0 {
		t.Fatal("confirm without preview")
	}
	send("/confirm-easy-athletics")
	if len(m.requests) != 0 {
		t.Fatal("base confirmation without preview")
	}
	send("/easy-athletics")
	if a.tacticalRequest == nil || a.tacticalRequest.Intent.EntryID != game.EasyAthleticsEntry || len(m.requests) != 0 {
		t.Fatal("base preview routing")
	}
	send("/confirm-tactical-mind")
	if len(m.requests) != 0 {
		t.Fatal("wrong confirmation command")
	}
	send("/confirm-easy-athletics")
	if a.tacticalRequest == nil || len(m.requests) != 1 {
		t.Fatal("base uncertain delivery lost identity")
	}
	m.fail = false
	send("/confirm-easy-athletics")
	if len(m.requests) != 2 || m.requests[0] != m.requests[1] || a.tacticalRequest != nil {
		t.Fatal("base retry identity")
	}
	m.requests = nil
	m.fail = true
	send("/tactical-mind")
	if a.tacticalRequest == nil || len(m.requests) != 0 || !strings.Contains(a.narrativeView().View(), "Preview only") {
		t.Fatal("preview unusable")
	}
	send("/confirm-tactical-mind")
	if a.tacticalRequest == nil || len(m.requests) != 1 {
		t.Fatal("lost identity")
	}
	m.fail = false
	send("/confirm-tactical-mind")
	if len(m.requests) != 2 || m.requests[0] != m.requests[1] || a.tacticalRequest != nil || len(m.inputs) != 0 {
		t.Fatal("retry changed identity or used inference")
	}
	send("/decline-tactical-mind")
	if a.tacticalRequest == nil || a.tacticalRequest.Intent.EntryID != "decline" {
		t.Fatal("decline routing")
	}
	send("/cancel-tactical-mind")
	if a.tacticalRequest != nil {
		t.Fatal("cancel failed")
	}
}
